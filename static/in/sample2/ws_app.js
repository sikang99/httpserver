var WebSocketServer = require('websocket').server;
var http = require('http');
//var fs = require('fs');
//var wav = require('wav');
var port = 9001;

var server = http.createServer(function(request, response) {
    console.log((new Date()) + ' Received request for ' + request.url);
    response.writeHead(404);
    response.end();
});
server.listen(port, function() {
    console.log((new Date()) + ' Server is listening on port '+port);
});
 
wsServer = new WebSocketServer({
    httpServer: server,
    // You should not use autoAcceptConnections for production 
    // applications, as it defeats all standard cross-origin protection 
    // facilities built into the protocol and the browser.  You should 
    // *always* verify the connection's origin and decide whether or not 
    // to accept it. 
    autoAcceptConnections: false
});
 
function originIsAllowed(origin) {
  // put logic here to detect whether the specified origin is allowed. 
  return true;
}

	
/*
var fileWriter = new wav.FileWriter('demo2.wav', {
	channels: 1,
	sampleRate: 48000,
	bitDepth: 16
});
*/

wsServer.on('request', function(request) {
    if (!originIsAllowed(request.origin)) {
      // Make sure we only accept requests from an allowed origin 
      request.reject();
      console.log((new Date()) + ' Connection from origin ' + request.origin + ' rejected.');
      return;
    }

    
    //var connection = request.accept('echo-protocol', request.origin);
	var connection = request.accept(null, request.origin);
    console.log((new Date()) + ' Connection accepted.' + request.origin);
	
	/*
	var fileWriter = new wav.FileWriter('demo2.wav', {
		channels: 1,
		sampleRate: 48000,
		bitDepth: 16
	});
	*/
	var total = 0;
    connection.on('message', function(message) {
        if (message.type === 'utf8') {
            console.log('Received Message: ' + message.utf8Data);
            connection.sendUTF(message.utf8Data);
        }
        else if (message.type === 'binary') {
			total += message.binaryData.length;
            console.log('Received Binary Message of ' +total+"/"+ message.binaryData.length + ' bytes');
            connection.sendBytes(message.binaryData);
        }
    });
    connection.on('close', function(reasonCode, description) {
        console.log((new Date()) + ' Peer ' + connection.remoteAddress + ' disconnected.');
    });
});