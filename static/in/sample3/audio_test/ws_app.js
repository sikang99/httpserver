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

var connect = null;
var connectionGet = null;
wsServer.on('request', function(request) {
    if (!originIsAllowed(request.origin)) {
      // Make sure we only accept requests from an allowed origin 
      request.reject();
      console.log((new Date()) + ' Connection from origin ' + request.origin + ' rejected.');
      return;
    }

    
    //var connection = request.accept('echo-protocol', request.origin);
	if (request.resource.indexOf("/get") >= 0) {
		connectionGet = request.accept(null, request.origin);
		console.log((new Date()) + ' Get Connection accepted.' + request.origin +":" + request.resource);
		
		connectionGet.on('message', function(message) {
			if (message.type === 'utf8') {
				console.log('Received Get Message: ' + message.utf8Data);
			}
		});
		connectionGet.on('close', function(reasonCode, description) {
			console.log((new Date()) + ' PeerGet ' + connection.remoteAddress + ' disconnected.');
		});
	} else {		
		connection = request.accept(null, request.origin);
		console.log((new Date()) + ' Post Connection accepted.' + request.origin +":" + request.resource);

		/*
		var fileWriter = new wav.FileWriter('demo2.wav', {
			channels: 1,
			sampleRate: 48000,
			bitDepth: 16
		});
		*/
		var total = 0;
		var header = null;
		var playerConn = null;
		connection.on('message', function(message) {
			if (message.type === 'utf8') {
				//console.log('Received Message: ' + message.utf8Data.length + ":"+ message.utf8Data);
				//console.log('Received Message: ' + message.utf8Data.length);
				if (message.utf8Data.indexOf("POST ") >= 0) {
					console.log('POST: ' + message.utf8Data);
				} else {
					//connection.sendUTF(message.utf8Data);
					//connection.sendUTF(new Float32Array(message.utf8Data));
					header = message.utf8Data;
				}
			}
			else if (message.type === 'binary') {
				total += message.binaryData.length;
				//console.log('Received Binary Message('+(typeof message.binaryData)+') of ' +total+"/"+ message.binaryData.length + ' bytes');
				var buf = new Buffer(header, "utf8");
				var sendBuf = Buffer.concat([buf, message.binaryData]);

				if (connectionGet != null)
				{
					connectionGet.sendBytes(sendBuf);
					//connectionGet.sendBytes(buf);
				}
			}
		});
		connection.on('close', function(reasonCode, description) {
			console.log((new Date()) + ' Peer ' + connection.remoteAddress + ' disconnected.');
		});
	}
});