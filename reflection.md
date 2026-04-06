# written comparison

## some prelimimary notes from lizzie (delete/polish up later! just writing down while it's still fresh)

1. just passed around the channels themselves. For the Go implementation, I mostly followed a very similar pattern to Elixir, where each process had one single channel that it used to receive messages, and even called it a mailbox. However, I ended up using two channels for the Barber process - one as the "general" channel, and another, dedicated channel to receive a rating from the customer who had just had their hair cut. This helped prevent interference from other messages, as the customer process was the only process to be passed that "rating" channel.
2. State management through local variables felt quite natural in Go, as this follows the pattern that many mainstream programming languages (e.g. Python) that we've learned use.
3. the Barber is achieved through two different functions in Go. The main loop `barber` handles all waking behavior that the barber has. The second function `sleepLoop` blocks the main loop. It just sits and waits for a wake up message, or a shutdown message. it passes a boolean value back to the main loop, so that the main loop knows whether to exit as a final shutdown, or to resume waking behavior.
4. my implementation of this was genuinely horrible - might go back and fix it tomorrow. i have one catch-all Message struct type with all possible fields. each message-sender only fills out the fields they need, and the receiver just has to trust that the right fields will be filled out. this is so unsafe and generally just bad.
5. n/a - no AI allowed for Go
