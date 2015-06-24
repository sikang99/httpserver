'use strict';

var Stream = (function stream() {
  function constructor(url) {
    this.url = url;
  }
  
  constructor.prototype = {
    readAll: function(progress, complete) {
	  var offset = 0;
      var xhr = new XMLHttpRequest();
      var async = true;
	  xhr.multipart = true;
      xhr.open("GET", this.url, async);
      xhr.responseType = "arraybuffer";
      if (progress) {
        xhr.onprogress = function (event) {
          progress(xhr.response, event.loaded, event.total);
        };
      }
      xhr.onreadystatechange = function (event) {

		//console.debug("readyState=" + xhr.readyState);
        if (xhr.readyState == 4) {
	 	  //var latestPart = xhr.responseText.substring(offset) 
    	  //offset = xhr.responseText.length;
		  //console.debug("offset=>" + offset);

          //console.debug("STATUS=" + xhr.status);
          complete(xhr.response);
          // var byteArray = new Uint8Array(xhr.response);
          // var array = Array.prototype.slice.apply(byteArray);
          // complete(array);
        }
      }
      xhr.send(null);
	
    }
  };
  return constructor;
})();


