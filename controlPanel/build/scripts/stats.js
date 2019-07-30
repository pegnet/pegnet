class GroupStatistics {
  constructor(stats) {
    this.stats = stats
  }

  totalHashPower() {
    let data = this.miners
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
    return acc
    return `${acc.toFixed(2).toLocaleString()} h/s`
  }
}