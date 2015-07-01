// variables
var recorder = null;
var recording = false;
var volume = null;
var audioInput = null;
var sampleRate = null;
var audioContext = null;
var context = null;
var outputElementPost = document.getElementById('outputPost');
var outputElementGet = document.getElementById('outputGet');
var outputString;

// feature detection 
if (!navigator.getUserMedia)
    navigator.getUserMedia = navigator.getUserMedia || navigator.webkitGetUserMedia ||
                  navigator.mozGetUserMedia || navigator.msGetUserMedia;

if (navigator.getUserMedia){
    navigator.getUserMedia({audio:true}, success, function(e) {
    alert('Error capturing audio.');
    });
} else alert('getUserMedia not supported in this browser.');


var source, buffer;
function success(e){
    // creates the audio context
    audioContext = window.AudioContext || window.webkitAudioContext;
    context = new audioContext();

	// we query the context sample rate (varies depending on platforms)
    sampleRate = context.sampleRate;

    console.log('succcess');
    
    // creates a gain node
    volume = context.createGain();

    // creates an audio node from the microphone incoming stream
    audioInput = context.createMediaStreamSource(e);

    // connect the stream to the gain node
    audioInput.connect(volume);

    /* From the spec: This value controls how frequently the audioprocess event is 
    dispatched and how many sample-frames need to be processed each call. 
    Lower values for buffer size will result in a lower (better) latency. 
    Higher values will be necessary to avoid audio breakup and glitches */
    //var bufferSize = 2048;
	var bufferSize = 2048;
    recorder = context.createScriptProcessor(bufferSize, 2, 2);


    recorder.onaudioprocess = function(e){
        if (!recording) return;
        var left = e.inputBuffer.getChannelData (0);
        var right = e.inputBuffer.getChannelData (1);

		sendToServer(left, right);
			
		/*
		// local playback
		var source = context.createBufferSource();
		source.buffer = e.inputBuffer;
		source.connect(context.destination);
		source.start(0);
		*/
		
    }

    // we connect the recorder
    volume.connect (recorder);
    recorder.connect (context.destination); 
}

var sendFlag = false;
var socket = null;
var host = "ws://localhost:9001/"
var size = 0;
function startSend() {
	// connect ws
	//socket = new WebSocket(host, "audio_stream");
	var host = $("#wsUrlPost").val();
	socket = new WebSocket(host);
	socket.binaryType = "arraybuffer";
	socket.onopen = function() {
		sendFlag = true;
		headerSent = false;
		console.log("socket.onopen");
	}

	socket.onmessage = function(msg) {
		if (typeof msg.data == "object") {
			size += msg.data.byteLength;
			console.log("socket.onmessage "+ typeof msg.data +" " + size +" "+ msg.data.byteLength);

			///////////////////////////////////////////////////////////////////////////////
			// playback using echo message
			/*
			var source = context.createBufferSource();
			var data = new Float32Array(msg.data);
			source.buffer = context.createBuffer(1, data.length, sampleRate);;
			source.buffer.copyToChannel(data, 0, 0);
			source.connect(context.destination);
			source.start(0);		
			*/
		}
	}

	socket.onclose = function() {
		sendFlag = false;
		console.log("socket.onclose");
	}

	//leftchannel.length = rightchannel.length = 0;
	outputElementPost.innerHTML = 'Sending now...';

	// set send flag true
	recording = true;
}

function stopSend() {
	// set send flag false
	recording = false;

	// disconnect ws
	socket.close();

    outputElementPost.innerHTML = 'Sending stopped...';
}

var sentTotal = 0;
var headerSent = false;
function sendToServer(left, right) {

	if (sendFlag == true)
	{
		var data = null;
		if (headerSent == false)
		{
			data = "POST /live?channel=100&source=1 HTTP/1.1\r\n"+
					"User-Agent: HeritTestApp\r\n"+
					"Content-Type: multipart/x-mixed-replace;boundary=--agilemedia\r\n";
			socket.send(data);
			headerSent = true;
		}

		data = "\r\n--agilemedia\r\n"+
				"Content-Type: audio/pcm\r\n"+
				"Content-Length: "+left.byteLength+"\r\n"+
				"X-Audio-Format: format=pcm_16; channel=1; frequency="+sampleRate+"\r\n\r\n";
		socket.send(data);
				
		data = left;
		socket.send(data);
		sentTotal += data.byteLength;
		outputElementPost.innerHTML = 'Sending now... ('+sentTotal+')';
        //console.log('sendToServer:' + data.byteLength +":"+sentTotal);
	}
}



var socketGet = null;
var getSize = 0;
function startGet() {
	// connect ws
	//socket = new WebSocket(host, "audio_stream");
	var host = $("#wsUrlGet").val();
	socketGet = new WebSocket(host);
	socketGet.binaryType = "arraybuffer";
	socketGet.onopen = function() {
		console.log("socketGet.onopen");
		var data = $("#wsGetStr").val();
		if (data.length > 0)
		{
			data += "\r\n\r\n";
			socketGet.send(data);
			console.log("Send get request:"+data);
			
		}
	}

	socketGet.onmessage = function(msg) {
		if (typeof msg.data == "object") {
			getSize += msg.data.byteLength;
			//console.log("socketGet.onmessage "+ typeof msg.data +" " + size +" "+ msg.data.byteLength);
			var headerStr = decodeUtf8(msg.data);
			var headers = getHeaders(headerStr);
			var audioFormat = getAudioFormat(headers["X-Audio-Format"]);
			//console.log("Get X-Audio-Format, SampleRate:"+ audioFormat["frequency"]);
			
			var audioBuf = msg.data.slice(headerStr.length);
			//console.log(decodeUtf8(msg.data));

			///////////////////////////////////////////////////////////////////////////////
			// playback using get session
			var source = context.createBufferSource();
			//var data = new Float32Array(msg.data);
			var data = new Float32Array(audioBuf);
			source.buffer = context.createBuffer(1, data.length, parseInt(audioFormat["frequency"]));;
			source.buffer.copyToChannel(data, 0, 0);
			source.connect(context.destination);
			source.start(0);
		
			outputElementGet.innerHTML = 'Receiving now... ('+getSize+')';
		}
	}

	socketGet.onclose = function() {
		console.log("socketGet.onclose");
	}

	outputElementGet.innerHTML = 'Receiving now...';
}

function stopGet() {
	// disconnect ws
	socketGet.close();

    outputElementGet.innerHTML = 'Receiving stopped...';
}

function getAudioFormat(val) {
	var format = [];
	var list = val.split(" ");
	for (var i=0; i<list.length; i++) {
		var tokens = list[i].split("=");
		if (tokens.length == 2) {
			format[tokens[0]] = tokens[1];
		}
	}
	return format;
}

function getHeaders(str) {
	var headers = [];
	var strs = str.split("\r\n");
	for (var i=0; i<strs.length; i++)
	{
		var header = strs[i];
		var pair = header.split(":");
		if (pair.length == 2) {
			headers[pair[0]] = pair[1];
		} else {
			//console.log("Invalid header:"+header);
		}
	}
	return headers;
}

function decodeUtf8(arrayBuffer) {
  var result = "";
  var i = 0;
  var c = 0;
  var c1 = 0;
  var c2 = 0;

  var data = new Uint8Array(arrayBuffer);

  // If we have a BOM skip it
  if (data.length >= 3 && data[0] === 0xef && data[1] === 0xbb && data[2] === 0xbf) {
    i = 3;
  }

  while (i < data.length) {
    c = data[i];

    if (c < 128) {
      result += String.fromCharCode(c);
	  if (data[i] == 13 && data[i+1] == 10 && data[i+2] == 13 && data[i+3] == 10)
	  {
		  result += "\n\r\n";
		  break;
	  }
      i++;
    } else {
		break;
	}
	/*else if (c > 191 && c < 224) {
      if( i+1 >= data.length ) {
        throw "UTF-8 Decode failed. Two byte character was truncated.";
      }
      c2 = data[i+1];
      result += String.fromCharCode( ((c&31)<<6) | (c2&63) );
      i += 2;
    } else {
      if (i+2 >= data.length) {
        throw "UTF-8 Decode failed. Multi byte character was truncated.";
      }
      c2 = data[i+1];
      c3 = data[i+2];
      result += String.fromCharCode( ((c&15)<<12) | ((c2&63)<<6) | (c3&63) );
      i += 3;
    }*/
  }
  return result;
}