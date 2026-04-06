# written comparison

## some prelimimary notes from lizzie (delete/polish up later! just writing down while it's still fresh)

1. **_Process Identity_**: For the Go implementation, we mostly followed a similar pattern to Elixir's mailboxes. For the most part, each process has one single channel (which we called a mailbox) that it uses to receive messages. These channels are created by the `ShopOwner` process for the Shop Owner, Barber, and the Waiting Room (lines `88-90`). Each Customer creates its own channel.
The Barber is the only process to use multiple channels. The first channel, created and passed in by the Shop Owner, serves as the "general" channel to receive all messages. The second channel `RatingReplyChannel` (created on line `276`) serves as a dedicated channel to receive a rating from the customer who has just had their hair cut. This helped prevent any possible interference from other messages, as the customer process was the only process to be passed that "rating" channel. When a message requires a response, such as a request for stats, the sender includes its own channel in the message struct.

2. **_State Management_**: State management through local variables felt quite natural in Go, because we are very familiar with tracking data changes inside functions. Most of our programming experience has been in languages that support this, such as Python.

3. **_Sleeping Barber Handshake_**: In Go, the Barber's behavior is achieved through two different functions. The main loop `barber` (`253-317`) handles all waking behavior that the barber has. When the Barber receives a command to sleep (`MsgNoneWaiting`), it calls the second function `barberSleepLoop` and waits for it to return. `barberSleepLoop` sits and waits for a wake up message, or a shutdown message. When it receives a message, it passes a value back to the main loop, so that the main loop knows whether to exit as a final shutdown, or to resume waking behavior. It also receives the current running averages, and handles requests for statistics, to prevent a locking scenario where the Shop Owner is stuck waiting forever for statistics from a sleeping barber.

4. _Message Types_ my implementation of this was genuinely horrible - might go back and fix it tomorrow. i have one catch-all Message struct type with all possible fields. each message-sender only fills out the fields they need, and the receiver just has to trust that the right fields will be filled out. this is so unsafe and generally just bad.

5. _Select vs. Receive_

6. _AI Tool Usage_
