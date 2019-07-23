package polling_test

import (
	"net/http"
	"testing"

	"github.com/pegnet/pegnet/common"
	. "github.com/pegnet/pegnet/polling"
	"github.com/zpatrick/go-config"
)

// TestFixedApiLayerPeggedAssets tests all the crypto assets are found on ApiLayer
func TestFixedApiLayerPeggedAssets(t *testing.T) {
	c := config.NewConfig([]config.Provider{common.NewUnitTestConfigProvider()})

	// Set default http client to return what we expect from apilayer
	cl := GetClientWithFixedResp([]byte(apiLayerReponse))
	http.DefaultClient = cl

	peg := make(PegAssets)
	APILayerInterface(c, peg)
	for _, asset := range common.CurrencyAssets {
		_, ok := peg[asset]
		if !ok {
			t.Errorf("Missing %s", asset)
		}
	}
}

var apiLayerReponse = `{"success":true,"terms":"https:\/\/currencylayer.com\/terms","privacy":"https:\/\/currencylayer.com\/privacy","timestamp":1563828666,"source":"USD","quotes":{"USDAED":3.67295,"USDAFN":79.850338,"USDALL":108.870258,"USDAMD":476.180354,"USDANG":1.78505,"USDAOA":346.555499,"USDARS":42.456988,"USDAUD":1.421297,"USDAWG":1.79975,"USDAZN":1.705017,"USDBAM":1.742903,"USDBBD":2.01975,"USDBDT":84.51296,"USDBGN":1.745099,"USDBHD":0.377005,"USDBIF":1851,"USDBMD":1,"USDBND":1.35065,"USDBOB":6.91285,"USDBRL":3.739797,"USDBSD":1.00025,"USDBTC":9.7194188e-5,"USDBTN":68.914102,"USDBWP":10.577987,"USDBYN":2.017397,"USDBYR":19600,"USDBZD":2.01575,"USDCAD":1.31165,"USDCDF":1665.499211,"USDCHF":0.9821,"USDCLF":0.02497,"USDCLP":689.000158,"USDCNY":6.881102,"USDCOP":3174.15,"USDCRC":574.735014,"USDCUC":1,"USDCUP":26.5,"USDCVE":98.222497,"USDCZK":22.786402,"USDDJF":177.719751,"USDDKK":6.66095,"USDDOP":51.044963,"USDDZD":119.260028,"USDEGP":16.619889,"USDERN":14.999739,"USDETB":29.004983,"USDEUR":0.892204,"USDFJD":2.1268,"USDFKP":0.80079,"USDGBP":0.8015,"USDGEL":2.875042,"USDGGP":0.801578,"USDGHS":5.389403,"USDGIP":0.80079,"USDGMD":49.964996,"USDGNF":9237.500707,"USDGTQ":7.665351,"USDGYD":208.824973,"USDHKD":7.81065,"USDHNL":24.650026,"USDHRK":6.591016,"USDHTG":94.061501,"USDHUF":290.269929,"USDIDR":13939,"USDILS":3.527506,"USDIMP":0.801578,"USDINR":68.929733,"USDIQD":1190,"USDIRR":42105.000124,"USDISK":124.819818,"USDJEP":0.801578,"USDJMD":134.619994,"USDJOD":0.707697,"USDJPY":107.87297,"USDKES":103.59735,"USDKGS":69.648033,"USDKHR":4082.999949,"USDKMF":438.800805,"USDKPW":900.064657,"USDKRW":1176.860062,"USDKWD":0.30425,"USDKYD":0.833405,"USDKZT":385.740062,"USDLAK":8684.999964,"USDLBP":1511.650177,"USDLKR":175.870017,"USDLRD":201.624975,"USDLSL":13.839987,"USDLTL":2.95274,"USDLVL":0.60489,"USDLYD":1.404995,"USDMAD":9.611502,"USDMDL":17.525499,"USDMGA":3607.499758,"USDMKD":54.663501,"USDMMK":1518.100677,"USDMNT":2664.879598,"USDMOP":8.04445,"USDMRO":357.000346,"USDMUR":35.898501,"USDMVR":15.449644,"USDMWK":760.054989,"USDMXN":19.055995,"USDMYR":4.112402,"USDMZN":61.814976,"USDNAD":13.839653,"USDNGN":360.000055,"USDNIO":33.50191,"USDNOK":8.61012,"USDNPR":110.244994,"USDNZD":1.478875,"USDOMR":0.38499,"USDPAB":1.00025,"USDPEN":3.285497,"USDPGK":3.389744,"USDPHP":51.109569,"USDPKR":159.669867,"USDPLN":3.78865,"USDPYG":5977.550021,"USDQAR":3.64175,"USDRON":4.212597,"USDRSD":105.029919,"USDRUB":63.102504,"USDRWF":910,"USDSAR":3.750502,"USDSBD":8.27135,"USDSCR":13.742494,"USDSDG":45.114502,"USDSEK":9.412501,"USDSGD":1.3609,"USDSHP":1.320899,"USDSLL":9250.000056,"USDSOS":580.000449,"USDSRD":7.458009,"USDSTD":21560.79,"USDSVC":8.75025,"USDSYP":515.000005,"USDSZL":13.839832,"USDTHB":30.849747,"USDTJS":9.430199,"USDTMT":3.5,"USDTND":2.86375,"USDTOP":2.267899,"USDTRY":5.679735,"USDTTD":6.77465,"USDTWD":31.072029,"USDTZS":2299.198699,"USDUAH":25.783993,"USDUGX":3695.203866,"USDUSD":1,"USDUYU":35.129707,"USDUZS":8630.000157,"USDVEF":9.987502,"USDVND":23227.5,"USDVUV":114.779918,"USDWST":2.602798,"USDXAF":584.559787,"USDXAG":0.061126,"USDXAU":0.000702,"USDXCD":2.70255,"USDXDR":0.7239,"USDXOF":591.501104,"USDXPF":106.697048,"USDYER":250.297342,"USDZAR":13.868106,"USDZMK":9001.20624,"USDZMW":12.825506,"USDZWL":322.000001}}`
