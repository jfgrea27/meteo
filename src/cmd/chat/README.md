# Weather Chat

The Weather Chat is a simple gRPC server that exposes NL to SQL using LLM.

## Run locally

```sh
# server side
just start chat
# client side
echo '{"text": "Hello"} {"text": "Weather in Paris?"}' | grpcurl -plaintext -d @ -import-path src/proto -proto weather_chat.proto localhost:50051 hello.WeatherChat/Chat
```
