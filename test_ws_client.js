const WebSocket = require('ws');

console.log('WebSocket client test starting...');

const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', function open() {
  console.log('WebSocket connected successfully!');

  // テストメッセージを送信
  const testMessage = {
    event: 'join-room',
    data: {
      roomId: 'TEST123',
      playerName: 'TestPlayer'
    }
  };

  console.log('Sending test message:', JSON.stringify(testMessage));
  ws.send(JSON.stringify(testMessage));
});

ws.on('message', function message(data) {
  console.log('Received message:', data.toString());
});

ws.on('error', function error(err) {
  console.error('WebSocket error:', err);
});

ws.on('close', function close(code, reason) {
  console.log('WebSocket closed:', code, reason.toString());
});

// 5秒後に接続を閉じる
setTimeout(() => {
  console.log('Closing connection...');
  ws.close();
}, 5000);
