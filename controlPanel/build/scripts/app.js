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
        $tr.find('td').eq(1).text(formatNumber(balance));

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
      $("#balance").val(formatNumber(data.balance));
  
    };


    //data: "auth_core.efficiency", render: function (data, type, row) {
    //                 return (data/100) + "%"
    //             }


    // Statistics table, only will show things when you are on the page? But better than nothing.
    stats = $('#mining-statistics').DataTable({
      columns: [
        {title: "ID", data: "id"},
        {title: "Block Height", data: "blockheight"},
        {title:"Tags", data:"tags", render: function (data, type, row) {
              return JSON.stringify(data)
          }
        },
        {title:"Miners", data:"miners", render: function (data, type, row) {
            return Object.keys(data).length
          }
        },
        {title:"Hash Power", data:"miners", render: function (data, type, row) {
            let k = Object.keys(data)

            let totalDur = moment.duration()
            let acc = 0
            for(let i = 0; i < k.length; i ++) {
              // TODO: follow format 2019-07-27T19:40:23.065954969-05:00
              let start = moment(data[k[i]].start)
              let stop = moment(data[k[i]].stop)
              let dur = moment.duration(stop.diff(start))
              acc = acc + (data[k[i]].totalhashes / dur.asSeconds())
              totalDur.add(dur)
            }
            return `${acc.toFixed(2).toLocaleString()} h/s`
          }
        }
      ],
      "bAutoWidth": false
    });

    statsEvents = new EventSource('/events/gstats');
    statsEvents.onmessage = function(event) {
      console.log(event)
      var data = JSON.parse(event.data);
      console.log(data)

      stats.rows.add([data]).draw()
    };


  });


})();

// For debugging
var stats

var d = JSON.parse(`{"miners":{"0":{"ID":0,"TotalHashes":56151,"BestDifficulty":18446695452556821197,"Start":"2019-07-27T19:05:08.192391284-05:00","Stop":"2019-07-27T19:05:20.153448502-05:00"},"1":{"ID":1,"TotalHashes":59196,"BestDifficulty":18446725004250009971,"Start":"2019-07-27T19:05:08.192399764-05:00","Stop":"2019-07-27T19:05:20.153438987-05:00"},"10":{"ID":10,"TotalHashes":56721,"BestDifficulty":18446360745704785966,"Start":"2019-07-27T19:05:08.240455129-05:00","Stop":"2019-07-27T19:05:20.15344892-05:00"},"11":{"ID":11,"TotalHashes":57648,"BestDifficulty":18446705486092468714,"Start":"2019-07-27T19:05:08.192427259-05:00","Stop":"2019-07-27T19:05:20.153385731-05:00"},"12":{"ID":12,"TotalHashes":59253,"BestDifficulty":18446600880769865823,"Start":"2019-07-27T19:05:08.208410791-05:00","Stop":"2019-07-27T19:05:20.153510572-05:00"},"13":{"ID":13,"TotalHashes":56175,"BestDifficulty":18446534717622478873,"Start":"2019-07-27T19:05:08.220329018-05:00","Stop":"2019-07-27T19:05:20.153474816-05:00"},"14":{"ID":14,"TotalHashes":56374,"BestDifficulty":18446380315606909278,"Start":"2019-07-27T19:05:08.192390329-05:00","Stop":"2019-07-27T19:05:20.153448888-05:00"},"2":{"ID":2,"TotalHashes":57186,"BestDifficulty":18446445259243670435,"Start":"2019-07-27T19:05:08.19238835-05:00","Stop":"2019-07-27T19:05:20.1535267-05:00"},"3":{"ID":3,"TotalHashes":55307,"BestDifficulty":18446664771463705622,"Start":"2019-07-27T19:05:08.220332355-05:00","Stop":"2019-07-27T19:05:20.153380705-05:00"},"4":{"ID":4,"TotalHashes":56038,"BestDifficulty":18446444764336979663,"Start":"2019-07-27T19:05:08.240456582-05:00","Stop":"2019-07-27T19:05:20.153470553-05:00"},"5":{"ID":5,"TotalHashes":55081,"BestDifficulty":18446569131831191410,"Start":"2019-07-27T19:05:08.192420358-05:00","Stop":"2019-07-27T19:05:20.153438799-05:00"},"6":{"ID":6,"TotalHashes":57126,"BestDifficulty":18446412691676508188,"Start":"2019-07-27T19:05:08.192433052-05:00","Stop":"2019-07-27T19:05:20.153502232-05:00"},"7":{"ID":7,"TotalHashes":58637,"BestDifficulty":18446107566532885558,"Start":"2019-07-27T19:05:08.220331556-05:00","Stop":"2019-07-27T19:05:20.153632793-05:00"},"8":{"ID":8,"TotalHashes":58392,"BestDifficulty":18446637020880195285,"Start":"2019-07-27T19:05:08.19240877-05:00","Stop":"2019-07-27T19:05:20.153531922-05:00"},"9":{"ID":9,"TotalHashes":57932,"BestDifficulty":18445246488631056733,"Start":"2019-07-27T19:05:08.220333021-05:00","Stop":"2019-07-27T19:05:20.153531206-05:00"}},"blockheight":828,"id":"Net-5577006791947779410","tags":{"src":"127.0.0.1:46444"}}
`)