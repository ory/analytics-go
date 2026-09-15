# SQA payload fixtures

The Identify, Track and Page fixtures describe the compact SQA format already
emitted by v5.0.1 (commit `156aadf`). That release's tests still expected the old
Segment format (`event`, `properties`, `userId`, `context.library`, etc.). Updating
those expectations fixes existing test failures; it does not change the wire
format. Memory fields such as `alloc` already serialize even when zero.

`test-enqueue-track.json` is the shared baseline for enqueue, interval flush,
timestamp, message-ID and multi-event batch tests. Tests vary only the event count
or explicit message ID. Enqueue already replaces the supplied timestamp with the
enqueue time. Expected payloads come from checked-in JSON, not the production
message serializer.

`compact-batch.json` covers a batch with all three supported event types and
populated metadata. The edition test checks plain and gzip delivery against this
baseline: an unset edition must be absent, and setting an edition may add only
`ed`. Existing fields and the batch envelope must remain unchanged.
