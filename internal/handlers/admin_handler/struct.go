package admin_handler

type Legend struct {
	Data      []string `json:"data"`
	Formatter string   `json:"formatter"`
	Padding   []int    `json:"padding"`
}

type Series struct {
	Data       []interface{} `json:"data"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	YAxisIndex int           `json:"yAxisIndex"`
}

type YAxisStruct struct {
	Data []string `json:"data,omitempty"`
	Max  string   `json:"max,omitempty"`
	Min  string   `json:"min,omitempty"`
	Type string   `json:"type"`
}

type EChartDataStruct struct {
	Legend    Legend   `json:"legend"`
	Series    []Series `json:"series"`
	TextStyle struct {
		FontSize int `json:"fontSize"`
	} `json:"textStyle"`
	Tooltip struct {
		AxisPointer struct {
			Type string `json:"type"`
		} `json:"axisPointer"`
		Trigger string `json:"trigger"`
	} `json:"tooltip"`
	XAxis struct {
		Data []string `json:"data"`
	} `json:"xAxis"`
	YAxis []YAxisStruct `json:"yAxis"`
}
