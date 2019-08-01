(function(){


  var request_api = function(data, cb) {
    $.ajax({
      url: 'http://localhost:8099/v1',
      data: JSON.stringify(data),
      type: 'POST',
      dataType: 'json',
      success: cb
    });
  }

  var request_oprs_by_height = function(height, cb) {
    request_api({'method': 'oprs-by-height', 'params': {'height': height}}, cb);
  }

  var request_balance = function(address, cb) {
    request_api({'method': 'balance', 'params': {'address': address}}, cb);
  }

  function formatNumber(num) {
    return num.toString().replace(/(\d)(?=(\d{3})+(?!\d))/g, '$1,')
  }


  $(document).ready(function(){


    $(".address-balance-form-js").on('submit', function(e) {
      e.preventDefault();
      var addr = $("#tpnt_address").val();
      request_balance(addr, function(resp){
        var balance = resp.result.balance;

        if (balance == undefined || balance < 0) {
          alert("address not found!");
          return;
        }

        var $tr = $('.address-balance-table-js tbody').find('tr[data-addr="'+addr+'"]');
        if($tr.length == 0){
          $tr = $('<tr><td>'+addr+'</td><td>0</td></tr>');
          $tr.attr('data-addr', addr);
          $('.address-balance-table-js tbody').append($tr);
        }
        $tr.find('td').eq(1).text(balance/1e8);

        $('.address-balance-table-js').removeClass('d-none');
        $("#tpnt_address").val("");
      });
    })


    e1 = new EventSource('/events/common');
    e1.onmessage = function(event) {
      var data = JSON.parse(event.data);
      $("#block_height").val(data.dbht);
      $("#current_minute").val(data.minute);
      $("#hash_rate").val(data.hashRate);
      $("#difficulty").val(data.difficulty);
      $("#balance").val(data.balance/1e8);
  
    };
  
  });

})();