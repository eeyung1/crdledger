package handler

import (
	"net/http"

	"crdledger/internal/service"
)

// TxRowData is what the "tx_row" template partial expects. It's used both
// when rendering the full transactions list and when HTMX asks for a
// single refreshed row (e.g. after mark-paid).
type TxRowData struct {
	Tx        service.TransactionView
	CSRFToken string
	Error     string
}

func buildTxRows(views []service.TransactionView, csrfToken string) []TxRowData {
	rows := make([]TxRowData, len(views))
	for i, v := range views {
		rows[i] = TxRowData{Tx: v, CSRFToken: csrfToken}
	}
	return rows
}

// NetBarChartData backs the "net_bar_list" partial — Variant picks which
// semantic color role the bars use ("positive" or "attention"), so the
// same markup renders both the "owed to you" and "you owe" charts.
type NetBarChartData struct {
	Rows    []service.CounterpartyNet
	Variant string
}

// isHTMXRequest reports whether the request came from an htmx-driven
// interaction (as opposed to a plain browser navigation / no-JS
// fallback), so handlers can choose between returning a full page and a
// small HTML fragment.
func isHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
