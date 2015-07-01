// variables
var recorder = null;
var recording = false;
var recordingLength = 0;
var volume = null;
var audioInput = null;
var sampleRate = null;
var audioContext = null;
var context = null;
var outputElement = document.getElementById('output');
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
	var host = $("#wsUrl").val();
	socket = new WebSocket(host);
	socket.binaryType = "arraybuffer";
	socket.onopen = function() {
		sendFlag = true;
		console.log("socket.onopen");
	}

	socket.onmessage = function(msg) {
		size += msg.data.byteLength;
		console.log("socket.onmessage " + size +" "+ msg.data.byteLength);
		
		///////////////////////////////////////////////////////////////////////////////
		// playback using echo message
		var source = context.createBufferSource();
		var data = new Float32Array(msg.data);
		source.buffer = context.createBuffer(1, data.length, sampleRate);;
		source.buffer.copyToChannel(data, 0, 0);
		source.connect(context.destination);
		source.start(0);
	}

	socket.onclose = function() {
		sendFlag = false;
		console.log("socket.onclose");
	}

	//leftchannel.length = rightchannel.length = 0;
	recordingLength = 0;
	outputElement.innerHTML = 'Sending now...';

	// set send flag true
	recording = true;
}

function stopSend() {
	// set send flag false
	recording = false;

	// disconnect ws
	socket.close();

    outputElement.innerHTML = 'Sending stopped...';
}

var sentTotal = 0;
function sendToServer(left, right) {

	if (sendFlag == true)
	{
		var data = left;
		socket.send(data);
		sentTotal += data.byteLength;
        console.log('sendToServer:' + data.byteLength +":"+sentTotal);
	}
}