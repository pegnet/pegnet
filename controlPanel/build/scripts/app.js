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
      var addr = $("#tpeg_address").val();
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
        $("#tpeg_address").val("");
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

    // Statistics table, only will show things when you are on the page? But better than nothing.
    stats = $('#mining-statistics').DataTable({
      order: [[ 0, "desc" ]],
      ajax:{
        url: "/cp/miningstats",
        dataSrc: function(d) {
          if(d.error != null) {
            console.log(d.error)
            return []
          }
          let stats = []
          let result = d.result
          for(var i = 0; i < result.length; i++) {
            var ks = Object.keys(result[i].allgroupstats)
            for(var k = 0; k < ks.length; k++) {
                  stats.push(result[i].allgroupstats[ks[k]])
            }
          }
          return stats
      }
      },
      columnDefs: [
        // { width: '40%', targets: 0 }
        { className: "mono-space", "targets": [ 6 ] }
      ],
      columns: [
        {title: "Block Height", data: "blockheight"},
        {title: "ID", data: "id"},
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
        },
        {title:"Hash Power Per Miner", data:"miners", render: function (data, type, row) {
            let k = Object.keys(data);

            let totalDur = moment.duration();
            let acc = 0
            for(let i = 0; i < k.length; i ++) {
              // TODO: follow format 2019-07-27T19:40:23.065954969-05:00
              let start = moment(data[k[i]].start);
              let stop = moment(data[k[i]].stop);
              let dur = moment.duration(stop.diff(start));
              acc = acc + dur.asSeconds() * (data[k[i]].totalhashes / dur.asSeconds());
              totalDur.add(dur)
            }
            acc = acc/totalDur.asSeconds();
            return `${acc.toFixed(2).toLocaleString()} h/s`
          }
        },
        {title: "Best Difficulty", data: "miners", render:function(data, type, row) {
            let best = 0;
            let k = Object.keys(data);
            console.log(data)
            for(let i = 0; i < k.length; i ++) {
              let m = data[k[i]];
              if(m.bestdifficulty > best) {
                best = m.bestdifficulty;
              }
              return best.toString(16);
            }
          }
        }
      ],
      "bAutoWidth": false
    });

    statsEvents = new EventSource('/events/gstats');
    statsEvents.onmessage = function(event) {
      console.log(event);
      var data = JSON.parse(event.data);
      console.log(data);

      stats.rows.add([data]).draw(false)
    };


  });

})();