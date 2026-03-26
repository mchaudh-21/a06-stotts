defmodule Config do
  # Configurable paramters live here
    def total_customers, do: 20
    def waiting_capacity, do: 5
    def arrival_min, do: 500
    def arrival_max, do: 2000
    def haircut_min, do: 1000
    def haircut_max, do: 4000
    def wait_threshold, do: 3000
    def shutdown_grace, do: 8000
end

defmodule Utils do
    def now(), do: System.monotonic_time(:millisecond)
    def rand_between(min, max) do
        :rand.uniform(max - min + 1) + min - 1
    end

    def log(who, msg) do
        IO.puts("[#{now()}] #{who}: #{msg}")
    end

    def clamp(min, max, val), do: max(min, min(max, val))
end

########################
# CUSTOMER PROCESS
########################
defmodule Customer do
    # Easch customer is its own process 
    def start(id, waiting_room) do
        spawn(fn ->
        arrival = Utils.now()
        send(waiting_room, {:arrive, self(), id, arrival})

        receive do
            :turned_away ->
            Utils.log("Customer #{id}", "turned away")
            :admitted ->
            wait_for_barber(id, arrival)
        end
        end)
    end

    defp wait_for_barber(id, arrival) do
        receive do
        {:start_cut, barber} ->
            Utils.log("Customer #{id}", "starting haircut")
            start_time = Utils.now()

            receive do
            {:rate_request, ^barber} ->
                wait = start_time - arrival
                jitter = Enum.random([-1, 0, 1])
                score =
                Utils.clamp(
                    1,
                    5,
                    5 - div(wait, Config.wait_threshold()) + jitter
                )

                Utils.log("Customer #{id}", "rating #{score} (wait #{wait}ms)")
                send(barber, {:rating, score})
            end
        end
    end
end

########################
# WAITING ROOM
########################
defmodule WaitingRoom do
    def start(barber) do
        spawn(fn -> loop([], 0, barber, false) end)
    end

    defp loop(queue, turned_away, barber, sleeping?) do
        receive do
        {:arrive, pid, id, arrival} ->
            if length(queue) >= Config.waiting_capacity() do
            send(pid, :turned_away)
            Utils.log("WaitingRoom", "turned away #{id}")
            loop(queue, turned_away + 1, barber, sleeping?)
            else
            send(pid, :admitted)
            Utils.log("WaitingRoom", "admitted #{id} (queue size #{length(queue)+1})") # IMPROVED
            new_queue = queue ++ [{pid, id, arrival}]

            # wake barber if sleeping
            if sleeping? do
                Utils.log("WaitingRoom", "waking barber") # ADDED
                send(barber, :wakeup)
            end

            loop(new_queue, turned_away, barber, false)
            end

        {:next_customer, barber_pid} ->
            case queue do
            [] ->
                send(barber_pid, :none_waiting)
                loop(queue, turned_away, barber, true)

            [{pid, id, arrival} | rest] ->
                send(barber_pid, {:customer_ready, pid, id, arrival})
                loop(rest, turned_away, barber, false)
            end

        {:get_stats, requester} ->
            send(requester, {:wr_stats, turned_away, length(queue)})
            loop(queue, turned_away, barber, sleeping?)

        :shutdown ->
            Utils.log("WaitingRoom", "shutdown")
        end
    end
end

########################
# BARBER
########################
defmodule Barber do
    def start() do
    spawn(fn -> loop(0, 0, 0, nil) end)
    end

    defp loop(count, total_time, total_rating, waiting_room) do
    receive do
        {:init, wr} ->
        send(wr, {:next_customer, self()})
        loop(count, total_time, total_rating, wr)

        :none_waiting ->
        Utils.log("Barber", "sleeping")
        sleep_loop(count, total_time, total_rating, waiting_room)

        {:customer_ready, pid, id, arrival} ->
        duration = Utils.rand_between(Config.haircut_min(), Config.haircut_max())
        Utils.log("Barber", "starting #{id} (#{duration}ms)")  # IMPROVED

        send(pid, {:start_cut, self()})

        :timer.sleep(duration)

        Utils.log("Barber", "finished haircut #{id}") # ADDED
        Utils.log("Barber", "requesting rating from #{id}") # ADDED

        send(pid, {:rate_request, self()})

        receive do
            {:rating, rating} ->
            Utils.log("Barber", "got rating #{rating}")
            new_count = count + 1
            new_time = total_time + duration
            new_rating = total_rating + rating

            send(waiting_room, {:next_customer, self()})
            loop(new_count, new_time, new_rating, waiting_room)
        end

        {:get_stats, requester} ->
        avg_time =
            if count > 0 do
            total_time / count
            else
            0
            end

        avg_rating =
            if count > 0 do
            total_rating / count
            else
            0
            end

        send(requester, {:barber_stats, count, avg_time, avg_rating})
        loop(count, total_time, total_rating, waiting_room)

        :shutdown ->
        Utils.log("Barber", "shutdown")
    end
  end

  defp sleep_loop(count, total_time, total_rating, wr) do
    receive do
      :wakeup ->
        Utils.log("Barber", "wakeup")
        send(wr, {:next_customer, self()})
        loop(count, total_time, total_rating, wr)

      :shutdown ->
        Utils.log("Barber", "shutdown")
    end
  end
end

########################
# SHOP OWNER
########################
defmodule ShopOwner do
    def start do
        barber = Barber.start()
        waiting_room = WaitingRoom.start(barber)

        send(barber, {:init, waiting_room})

        spawn_customers(waiting_room, Config.total_customers())

        :timer.sleep(Config.shutdown_grace())

        send(waiting_room, {:get_stats, self()})
        send(barber, {:get_stats, self()})

        wr_stats =
        receive do
            {:wr_stats, turned, _} -> turned
        end

        {served, avg_time, avg_rating} =
        receive do
            {:barber_stats, c, t, r} -> {c, t, r}
        end

        IO.puts("=== Barbershop Closing Report ===")
        IO.puts("Total customers arrived: #{Config.total_customers()}") # FIXED wording
        IO.puts("Customers served: #{served}")
        IO.puts("Customers turned away: #{wr_stats}")
        IO.puts("Average haircut duration: #{Float.round(avg_time / 1000, 2)}s")
        IO.puts("Average satisfaction: #{Float.round(avg_rating, 2)}")
        IO.puts("=================================")

        send(barber, :shutdown)
        send(waiting_room, :shutdown)

        :timer.sleep(200) # ADDED (ensure logs print)
    end

    defp spawn_customers(_, 0), do: :ok

    defp spawn_customers(wr, n) do
        Customer.start(n, wr)
        :timer.sleep(Utils.rand_between(Config.arrival_min(), Config.arrival_max()))
        spawn_customers(wr, n - 1)
    end
end

ShopOwner.start()