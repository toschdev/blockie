// Copyright 2018 Goole Inc.
// Copyright 2024 Tobias Schwarz
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Binary explorer demo. Displays widgets for insights of blockchain behaviour.
// Exist when 'q' or 'esc' is pressed.
package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/resty.v1"

	"github.com/sacOO7/gowebsocket"
	"github.com/tidwall/gjson"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/text"

	"github.com/ignite/cli/v28/ignite/services/plugin"
	"github.com/tidwall/pretty"

	"github.com/jedib0t/go-pretty/v6/list"
)

// optional port variable. example: `gex -p 30057`
var givenPort = flag.Int("p", 26657, "port to connect")
var givenHost = flag.String("h", "localhost", "host to connect")
var ssl = flag.Bool("s", false, "use SSL for connection")

// Info describes a list of types with data that are used in the explorer
type Info struct {
	blocks       *Blocks
	transactions *Transactions
}

// Blocks describe content that gets parsed for block
type Blocks struct {
	amount               int
	secondsPassed        int
	totalGasWanted       int64
	gasWantedLatestBlock int64
	maxGasWanted         int64
	lastTx               int64
}

// Transactions describe content that gets parsed for transactions
type Transactions struct {
	amount uint64
}

// playType indicates the type of the donut widget.
type playType int

func ShowBlockie(ctxApp context.Context, cp *plugin.ExecutedCommand) error {
	flag.Parse()

	// Init internal variables
	info := Info{}
	info.blocks = new(Blocks)
	info.transactions = new(Transactions)

	connectionSignal := make(chan string)

	networkInfo, err := getFromRPC("status")
	if err != nil {
		fmt.Println("Application not running on " + fmt.Sprintf("%s:%d", *givenHost, *givenPort))
		fmt.Println(err)
		os.Exit(1)
	}

	networkStatus := gjson.Parse(networkInfo)

	genesisRPC, _ := getFromRPC("genesis")

	genesisInfo := gjson.Parse(genesisRPC)

	ctx, cancel := context.WithCancel(context.Background())

	// START INITIALISING WIDGETS

	// Creates Network Widget
	currentNetworkWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := currentNetworkWidget.Write(networkStatus.Get("result.node_info.network").String()); err != nil {
		panic(err)
	}

	// Creates Health Widget
	healthWidget, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := healthWidget.Write("⌛ loading"); err != nil {
		panic(err)
	}

	// Creates System Time Widget
	timeWidget, err := text.New()
	if err != nil {
		panic(err)
	}
	currentTime := time.Now()
	if err := timeWidget.Write(fmt.Sprintf("%s\n", currentTime.Format("2006-01-02\n03:04:05 PM"))); err != nil {
		panic(err)
	}

	// Creates Connected Peers Widget
	peerWidget, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := peerWidget.Write("0"); err != nil {
		panic(err)
	}

	// Creates Seconds Between Blocks Widget
	secondsPerBlockWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := secondsPerBlockWidget.Write("0"); err != nil {
		panic(err)
	}

	// Creates Max Block Size Widget
	maxBlocksizeWidget, err := text.New()

	consensusParamsRPC, _ := getFromRPC("consensus_params")

	maxBlockSize := gjson.Get(consensusParamsRPC, "result.consensus_params.block.max_bytes").Int()
	if err != nil {
		panic(err)
	}
	if err := maxBlocksizeWidget.Write(fmt.Sprintf("%s", byteCountDecimal(maxBlockSize))); err != nil {
		panic(err)
	}

	// Creates Validators widget
	validatorWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := validatorWidget.Write("List available validators.\n\n"); err != nil {
		panic(err)
	}

	// Creates Validators widget
	gasMaxWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := gasMaxWidget.Write(" "); err != nil {
		panic(err)
	}

	// Creates Gas per Average Block Widget
	gasAvgBlockWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := gasAvgBlockWidget.Write("How much gas.\n\n"); err != nil {
		panic(err)
	}

	// Creates Gas per Average Transaction Widget
	gasAvgTransactionWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := gasAvgTransactionWidget.Write("How much gas.\n\n"); err != nil {
		panic(err)
	}

	// Creates Gas per Latest Transaction Widget
	latestGasWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := latestGasWidget.Write("How much gas.\n\n"); err != nil {
		panic(err)
	}

	// Transaction parsing widget
	transactionWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := transactionWidget.Write("Transactions will appear as soon as they are confirmed in a block.\n\n"); err != nil {
		panic(err)
	}

	// Create Blocks parsing widget
	blocksWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := blocksWidget.Write(" "); err != nil {
		panic(err)
	}

	// Complete Block parsing widget
	completeBlocksWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := completeBlocksWidget.Write("Complete Block as soon as confirmed.\n\n"); err != nil {
		panic(err)
	}

	// Complete Block parsing widget
	blocksHeaderWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := blocksHeaderWidget.Write("Block Header.\n\n"); err != nil {
		panic(err)
	}

	// Complete Block parsing widget
	blocksTreeWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := blocksTreeWidget.Write("Waiting for first block...\n\n"); err != nil {
		panic(err)
	}

	// Creates Gas per Average Transaction Widget
	blockieWidget, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		panic(err)
	}
	if err := blockieWidget.Write("\n" +
		"    BLOCKIE\n" +
		"   +------+\n" +
		"  /      /|\n" +
		" +------+ |\n" +
		" |  ..  | +\n" +
		" | \\__/ |/\n" +
		" +------+\n"); err != nil {
		panic(err)
	}

	// END INITIALISING WIDGETS

	// The functions that execute the updating widgets.

	// system powered widgets
	go writeTime(ctx, info, timeWidget, 1*time.Second)

	go writeSecondsPerBlock(ctx, info, secondsPerBlockWidget, 1*time.Second)
	go writeGasWidget(ctx, info, gasMaxWidget, gasAvgBlockWidget, gasAvgTransactionWidget, latestGasWidget, 1000*time.Millisecond, connectionSignal, genesisInfo)

	// websocket powered widgets
	go writeBlocks(ctx, info, blocksWidget, completeBlocksWidget, blocksHeaderWidget, blocksTreeWidget, connectionSignal)
	// go writeCompleteBlocks(ctx, info, completeBlocksWidget, blocksHeaderWidget, connectionSignal)
	go writeTransactions(ctx, info, transactionWidget, connectionSignal)

	t, err := termbox.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	// Draw Dashboard
	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("BLOCKIE: PRESS Q or ESC TO QUIT"),
		container.BorderColor(cell.ColorNumber(2)),
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.SplitVertical(
							container.Left(
								container.SplitVertical(
									container.Left(

										container.Border(linestyle.Light),
										container.BorderTitle("Chain-ID"),
										container.PlaceWidget(currentNetworkWidget),
									),
									container.Right(
										container.Border(linestyle.Light),
										container.BorderTitle("Max Block Size"),
										container.PlaceWidget(maxBlocksizeWidget),
									),
								),
							),
							container.Right(
								container.SplitVertical(
									container.Left(
										container.Border(linestyle.Light),
										container.BorderTitle("Gas Max"),
										container.PlaceWidget(gasMaxWidget),
									),
									container.Right(
										container.Border(linestyle.Light),
										container.BorderTitle("timeout_commit"),
										container.PlaceWidget(secondsPerBlockWidget),
									),
								),
							),
						),
					),
					container.Bottom(
						container.SplitVertical(
							container.Left(
								container.Border(linestyle.Light),
								container.BorderTitle("Latest Height"),
								container.AlignVertical(align.VerticalMiddle),
								container.PlaceWidget(blocksWidget),
								container.PaddingLeftPercent(15),
							),
							container.Right(
								container.Border(linestyle.Light),
								container.BorderTitle("Block Overview"),
								container.PlaceWidget(blocksTreeWidget),
							),
						),
					),
					container.SplitPercent(40),
				),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.SplitVertical(
							container.Left(
								container.Border(linestyle.Light),
								container.BorderTitle("Block Header"),
								container.PlaceWidget(blocksHeaderWidget),
							),
							container.Right(
								container.Border(linestyle.Light),
								container.PlaceWidget(blockieWidget),
							),
							container.SplitPercent(90),
						),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.BorderTitle("Latest Block in Detail"),
						container.PlaceWidget(completeBlocksWidget),
					),
					container.SplitPercent(40),
				),
			),
			container.SplitPercent(30),
		),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' || k.Key == keyboard.KeyEsc {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}

	return nil
}

// writeGasWidget writes the status to the healthWidget.
// Exits when the context expires.
func writeGasWidget(ctx context.Context, info Info, tMax *text.Text, tAvgBlock *text.Text, tAvgTx *text.Text, tLatest *text.Text, delay time.Duration, connectionSignal chan string, genesisInfo gjson.Result) {
	tMax.Reset()

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tMax.Reset()

			if info.blocks.maxGasWanted == 0 {
				tMax.Write(fmt.Sprintf("%v", "unlimited"))
			} else {
				// tMax.Write(fmt.Sprintf("%v", numberWithComma(info.blocks.maxGasWanted)))
				tMax.Write(fmt.Sprintf("%v", info.blocks.maxGasWanted))
			}
			
		case <-ctx.Done():
			return
		}
	}
}

// writeSecondsPerBlock writes the status to the Time per block.
// Exits when the context expires.
func writeSecondsPerBlock(ctx context.Context, info Info, t *text.Text, delay time.Duration) {

	t.Reset()

	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.Reset()
			blocksPerSecond := 0.00
			if info.blocks.secondsPassed != 0 {
				blocksPerSecond = float64(info.blocks.secondsPassed) / float64(info.blocks.amount)
			}

			t.Write(fmt.Sprintf("%.2f", blocksPerSecond))
		case <-ctx.Done():
			return
		}
	}
}

// WEBSOCKET WIDGETS

// writeTime writes the current system time to the timeWidget.
// Exits when the context expires.
func writeTime(ctx context.Context, info Info, t *text.Text, delay time.Duration) {
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			currentTime := time.Now()
			t.Reset()
			if err := t.Write(fmt.Sprintf("%s\n", currentTime.Format("2006-01-02\n03:04:05 PM"))); err != nil {
				panic(err)
			}
			info.blocks.secondsPassed++
		case <-ctx.Done():
			return
		}
	}
}

// writeBlocks writes the latest Block to the blocksWidget.
// Exits when the context expires.
func writeBlocks(ctx context.Context, info Info, t *text.Text, tCompleteBlock *text.Text, tHeader *text.Text, tBlocksTree *text.Text, connectionSignal <-chan string) {
	socket := gowebsocket.New(getWsUrl() + "/websocket")
	t.Reset()

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		currentBlock := gjson.Get(message, "result.data.value.block.header.height")
		currentFullBlock := gjson.Get(message, "result.data.value")
		if currentBlock.String() != "" {

			currentBlockHeight := numberWithComma(int64(currentBlock.Int()))
			// Populate latest block
			err := t.Write(fmt.Sprintf("%v\n", currentBlockHeight))
			if err != nil {
				panic(err)
			}

			info.blocks.amount++
			block_max_gas := gjson.Get(message, "result.data.value.result_end_block.consensus_param_updates.block.max_gas")
			info.blocks.maxGasWanted = block_max_gas.Int()

			// Populate Header
			tHeader.Reset()
			colorOutputHeader := pretty.Pretty([]byte(currentFullBlock.Get("block.header").Raw))
			if err := tHeader.Write(fmt.Sprintf("%s\n", colorOutputHeader)); err != nil {
				panic(err)
			}

			// Populate Bigger Block Details
			tCompleteBlock.Reset()
			tBlocksTree.Reset()

			colorOutput := pretty.Pretty([]byte(currentFullBlock.Raw))
			if err := tCompleteBlock.Write(fmt.Sprintf("%s\n", colorOutput)); err != nil {
				panic(err)
			}

			// Populate BlocksTree

			l := list.NewWriter()
			lTemp := list.List{}
			l.SetStyle(list.StyleConnectedRounded)
			lTemp.Render() // just to avoid the compile error of not using the object

			// First depth
			currentFullBlock.ForEach(func(key, value gjson.Result) bool {
				l.AppendItem(key)

				if key.String() == "result_finalize_block" {
					return true // keep iterating
				}
				// Second depth
				l.Indent()
				value.ForEach(func(key_sub, value_sub gjson.Result) bool {
					l.AppendItem(key_sub)

					// Third depth
					l.Indent()
					value_sub.ForEach(func(key_sub2, value_sub2 gjson.Result) bool {
						l.AppendItem(key_sub2)

						return true // keep iterating
					})
					l.UnIndent()

					return true // keep iterating
				})
				l.UnIndent()

				return true // keep iterating
			})

			for _, line := range strings.Split(l.Render(), "\n") {
				if err := tBlocksTree.Write(fmt.Sprintf("%s\n", line)); err != nil {
					panic(err)
				}
			}

		}

	}

	socket.Connect()

	socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event='NewBlock'\"], \"id\": 1 }")

	for {
		select {
		case s := <-connectionSignal:
			if s == "no_connection" {
				socket.Close()
			}
			if s == "reconnect" {
				writeBlocks(ctx, info, t, tCompleteBlock, tHeader, tBlocksTree, connectionSignal)
			}
		case <-ctx.Done():
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}

// writeBlockDonut continuously changes the displayed percent value on the donut by the
// step once every delay. Exits when the context expires.
func writeBlockDonut(ctx context.Context, d *donut.Donut, start, step int, delay time.Duration, pt playType, connectionSignal <-chan string) {
	socket := gowebsocket.New(getWsUrl() + "/websocket")

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		step := gjson.Get(message, "result.data.value.step")
		progress := 0

		if step.String() == "RoundStepNewHeight" {
			progress = 100
		}

		if step.String() == "RoundStepCommit" {
			progress = 80
		}

		if step.String() == "RoundStepPrecommit" {
			progress = 60
		}

		if step.String() == "RoundStepPrevote" {
			progress = 40
		}

		if step.String() == "RoundStepPropose" {
			progress = 20
		}

		if err := d.Percent(progress); err != nil {
			panic(err)
		}

	}

	socket.Connect()

	socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event='NewRoundStep'\"], \"id\": 3 }")

	for {
		select {
		case s := <-connectionSignal:
			if s == "no_connection" {
				socket.Close()
			}
			if s == "reconnect" {
				writeBlockDonut(ctx, d, start, step, delay, pt, connectionSignal)
			}
		case <-ctx.Done():
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}

// writeTransactions writes the latest Transactions to the transactionsWidget.
// Exits when the context expires.
func writeTransactions(ctx context.Context, info Info, t *text.Text, connectionSignal <-chan string) {
	socket := gowebsocket.New(getWsUrl() + "/websocket")

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		currentTx := gjson.Get(message, "result.data.value.TxResult.result.log")
		currentTime := time.Now()
		if currentTx.String() != "" {
			if err := t.Write(fmt.Sprintf("%s\n", currentTime.Format("2006-01-02 03:04:05 PM")+"\n"+currentTx.String())); err != nil {
				panic(err)
			}

			info.blocks.totalGasWanted = info.blocks.totalGasWanted + gjson.Get(message, "result.data.value.TxResult.result.gas_wanted").Int()
			info.blocks.lastTx = gjson.Get(message, "result.data.value.TxResult.result.gas_wanted").Int()
			info.transactions.amount++
		}
	}

	socket.Connect()

	socket.SendText("{ \"jsonrpc\": \"2.0\", \"method\": \"subscribe\", \"params\": [\"tm.event='Tx'\"], \"id\": 2 }")

	for {
		select {
		case s := <-connectionSignal:
			if s == "no_connection" {
				socket.Close()
			}
			if s == "reconnect" {
				writeTransactions(ctx, info, t, connectionSignal)
			}
		case <-ctx.Done():
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}

// UTIL FUNCTIONS

// Get Data from RPC Endpoint
func getFromRPC(endpoint string) (string, error) {
	resp, err := resty.R().
		SetHeader("Cache-Control", "no-cache").
		SetHeader("Content-Type", "application/json").
		Get(getHttpUrl() + "/" + endpoint)

	return resp.String(), err
}

// byteCountDecimal calculates bytes integer to a human readable decimal number
func byteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func numberWithComma(n int64) string {
	in := strconv.FormatInt(n, 10)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

func getHttpUrl() string {
	return getUrl("http", *ssl)
}

func getWsUrl() string {
	return getUrl("ws", *ssl)
}

func getUrl(protocol string, secure bool) string {
	if secure {
		protocol = protocol + "s"
	}

	return fmt.Sprintf("%s://%s:%d", protocol, *givenHost, *givenPort)
}
