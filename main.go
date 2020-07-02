package main

import (
	"market-cli/chart"
	"time"

	"github.com/piquette/finance-go/datetime"
	"github.com/spf13/cobra"
)

var (
	name       *string
	year       *bool
	oneDay     *bool
	fiveDays   *bool
	oneMonth   *bool
	threeMonth *bool
	sixMonth   *bool
	ytd        *bool
	oneYear    *bool
	twoYears   *bool
	fiveYears  *bool
	max        *bool
)

func main() {

	rootCmd := &cobra.Command{
		Use:   "market",
		Short: "Display stocks in realtime",
		Run: func(cmd *cobra.Command, args []string) {
			symbol := args[0]
			interval, startTime := parseTimeRange()

			stock := chart.Stock{
				Symbol:    symbol,
				Interval:  interval,
				StartTime: startTime,
			}
			chart.PlayChart(stock)
		},
	}

	oneDay = rootCmd.PersistentFlags().Bool("1d", false, "Current day")
	fiveDays = rootCmd.PersistentFlags().Bool("5d", false, "Last 5 days")
	oneMonth = rootCmd.PersistentFlags().Bool("1m", false, "Last 1 month")
	threeMonth = rootCmd.PersistentFlags().Bool("3m", false, "Last 3 months")
	sixMonth = rootCmd.PersistentFlags().Bool("6m", false, "Last 6 months")
	ytd = rootCmd.PersistentFlags().Bool("ytd", false, "Year to date")
	oneYear = rootCmd.PersistentFlags().Bool("1y", false, "Last 1 year")
	twoYears = rootCmd.PersistentFlags().Bool("2y", false, "Last 2 year")
	fiveYears = rootCmd.PersistentFlags().Bool("5y", false, "Last 5 year")
	max = rootCmd.PersistentFlags().Bool("max", false, "Display whole history")

	rootCmd.Execute()
}

func parseTimeRange() (datetime.Interval, time.Time) {
	switch {
	case *oneDay:
		return datetime.FiveMins, time.Now().AddDate(0, 0, -1)
	case *fiveDays:
		return datetime.NinetyMins, time.Now().AddDate(0, 0, -5)
	case *oneMonth:
		return datetime.OneDay, time.Now().AddDate(0, -1, 0)
	case *threeMonth:
		return datetime.OneDay, time.Now().AddDate(0, -3, 0)
	case *sixMonth:
		return datetime.OneDay, time.Now().AddDate(0, -6, 0)
	case *ytd:
		year, _, _ := time.Now().Date()
		currentLocation := time.Now().Location()
		return datetime.OneDay, time.Date(year, 1, 1, 0, 0, 0, 0, currentLocation)
	case *oneYear:
		return datetime.FiveDay, time.Now().AddDate(-1, 0, 0)
	case *twoYears:
		return datetime.OneMonth, time.Now().AddDate(-2, 0, 0)
	case *fiveYears:
		return datetime.OneMonth, time.Now().AddDate(-5, 0, 0)
	case *max:
		return datetime.OneMonth, time.Now().AddDate(-100, 0, 0)
	default:
		panic("no time range was found")
	}
}
