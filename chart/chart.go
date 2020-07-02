package chart

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	"github.com/shopspring/decimal"
)

type Stock struct {
	Symbol    string
	Interval  datetime.Interval
	StartTime time.Time
}

type stockMetadata struct {
	symbol             string
	chartPreviousClose decimal.Decimal
	currency           string
	lastPrice          decimal.Decimal
	lastPriceTime      *datetime.Datetime
	diff               decimal.Decimal
	diffPrc            decimal.Decimal
	dateRange          string
	interval           string
	exchange           string
}

func PlayChart(stock Stock) {
	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	header, err := text.New()
	if err != nil {
		panic(err)
	}

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
	go playLineChart(stock, ctx, lc, header)
	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.PlaceWidget(header),
			),
			container.Bottom(
				container.PlaceWidget(lc),
			),
			container.SplitFixed(3),
		),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}

func playLineChart(stock Stock, ctx context.Context,
	lc *linechart.LineChart, header *text.Text) {

	now := time.Now()
	params := &chart.Params{Symbol: stock.Symbol, Interval: stock.Interval,
		Start: datetime.New(&stock.StartTime), End: datetime.New(&now)}
	q := chart.Get(params)

	data := []float64{}
	xLabels := map[int]string{}
	format := pickDateFormat(*params.End, *params.Start)

	var meta *stockMetadata
	for i := 0; q.Next(); i++ {
		if meta == nil {
			meta = &stockMetadata{
				symbol:             stock.Symbol,
				chartPreviousClose: decimal.NewFromFloat(q.Meta().ChartPreviousClose),
				currency:           q.Meta().Currency,
				interval:           q.Meta().DataGranularity,
				exchange:           q.Meta().ExchangeName,
			}
		}

		bar := q.Bar()
		if bar.Close.IsZero() {
			continue
		}
		floatData, _ := bar.Close.Float64()
		data = append(data, floatData)
		xLabel := datetime.FromUnix(bar.Timestamp)
		xLabels[i] = xLabel.Time().Format(format)

		meta.lastPrice = bar.Close
		meta.lastPriceTime = datetime.FromUnix(q.Bar().Timestamp)
	}
	meta.diff = meta.lastPrice.Sub(meta.chartPreviousClose)
	meta.diffPrc = meta.diff.Div(meta.chartPreviousClose).Mul(decimal.NewFromInt(100))

	setHeaderText(meta, header)

	if err := lc.Series("main", data,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorWhite)),
		linechart.SeriesXLabels(xLabels),
	); err != nil {
		panic(err)
	}
}

func setHeaderText(meta *stockMetadata, header *text.Text) {
	var txt string
	magenta := text.WriteCellOpts(cell.FgColor(cell.ColorMagenta))

	var diffColor text.WriteOption
	if meta.diff.IsPositive() {
		diffColor = text.WriteCellOpts(cell.FgColor(cell.ColorGreen))
	} else {
		diffColor = text.WriteCellOpts(cell.FgColor(cell.ColorRed))
	}
	txt = fmt.Sprintf("%s %s", meta.symbol, meta.currency)
	header.Write(txt, magenta)
	header.Write(" | Prev close at ")
	header.Write(meta.chartPreviousClose.StringFixed(2), magenta)
	header.Write(" | ")

	header.Write("Current ")
	header.Write(meta.lastPrice.StringFixed(2), magenta)
	header.Write(fmt.Sprintf(" %s(%s)", meta.diff.StringFixed(2), meta.diffPrc.StringFixed(2)), diffColor)
	header.Write(fmt.Sprintf(" on %s", meta.lastPriceTime.Time().Format("02/01/2006 15:04")))
	header.Write(" | ")
	header.Write(meta.exchange)
}

func pickDateFormat(end datetime.Datetime, start datetime.Datetime) string {
	if end.Unix()-start.Unix() > 86400 {
		return "2 Jan"
	}
	return "15:04"
}
