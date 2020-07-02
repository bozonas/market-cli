package chart

import (
	"context"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
)

type Stock struct {
	Symbol    string
	Interval  datetime.Interval
	StartTime time.Time
}

func PlayChart(stock Stock) {
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	const redrawInterval = 250 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	lc, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorCyan)),
		linechart.YAxisAdaptive(),
	)
	if err != nil {
		panic(err)
	}
	go playLineChart(stock, ctx, lc, redrawInterval/3)
	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(lc),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}
}

func playLineChart(stock Stock, ctx context.Context,
	lc *linechart.LineChart, delay time.Duration) {

	now := time.Now()
	params := &chart.Params{Symbol: stock.Symbol, Interval: stock.Interval,
		Start: datetime.New(&stock.StartTime), End: datetime.New(&now), IncludeExt: false}
	q := chart.Get(params)
	data := []float64{}
	xLabels := map[int]string{}
	format := pickDateFormat(*params.End, *params.Start)

	for i := 0; q.Next(); i++ {
		bar := q.Bar()
		if bar.Close.IsZero() {
			continue
		}
		floatData, _ := bar.Close.Float64()
		data = append(data, floatData)
		xLabel := datetime.FromUnix(bar.Timestamp)
		xLabels[i] = xLabel.Time().Format(format)
	}

	if err := lc.Series("main", data,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorWhite)),
		linechart.SeriesXLabels(xLabels),
	); err != nil {
		panic(err)
	}
}

func pickDateFormat(end datetime.Datetime, start datetime.Datetime) string {
	if end.Unix()-start.Unix() > 86400 {
		return "2 Jan"
	}
	return "15:04"
}
