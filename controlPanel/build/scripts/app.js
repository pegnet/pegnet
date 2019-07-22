(function(){

    e1 = new EventSource('/events/common');
    e1.onmessage = function(event) {
      var data = JSON.parse(event.data);
      $("#block_height").val(data.dbht);
      $("#current_minute").val(data.minute);
    };



})();