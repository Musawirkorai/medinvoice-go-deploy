package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
)

type Item struct {
	Name  string
	Qty   int
	Price float64
	Total float64
}

type Invoice struct {
	StoreName  string
	DoctorName string
	UserName   string
	InvoiceNo  string
	Date       string
	Items      []Item
	TotalBill  float64
}

func main() {

	//---------------------------------------
	// RAILWAY PORT FIX (REQUIRED)
	//---------------------------------------
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // for local runs
	}

	//---------------------------------------
	// MAIN PAGE TEMPLATE
	//---------------------------------------
	tpl := template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>Medical Invoice System</title>
	<style>
		body { font-family: 'Segoe UI'; background: #eaf4ff; margin: 0; padding: 0; }
		.container { width: 80%; margin: 40px auto; background: #fff; padding: 30px;
					 box-shadow: 0 0 20px rgba(0,0,0,0.15); border-radius: 10px; }
		h1 { text-align: center; color: #1e4e79; }
		label { font-weight: bold; color: #333; }
		input[type=text], input[type=number] {
			width: 100%; padding: 8px; margin: 5px 0;
			border-radius: 5px; border: 1px solid #a9c3d9;
		}
		table { width: 100%; border-collapse: collapse; margin-top: 20px; }
		th, td { padding: 12px; border-bottom: 1px solid #ddd; }
		th { background-color: #1e88e5; color: white; }
		tr:nth-child(even) { background-color: #f9f9f9; }
		button {
			background-color: #1e88e5; color: white; border: none; 
			padding: 12px 20px; margin-top: 20px; cursor: pointer;
			border-radius: 5px; font-size: 16px;
		}
		button:hover { background-color: #0d6cd1; }
	</style>
</head>
<body>

<div class="container">
	<h1>Medical Billing System</h1>
	<form id="invoiceForm" target="_blank" method="POST" action="/generate">

		<label>Patient Name:</label>
		<input type="text" name="username" placeholder="Enter patient name" required>

		<label>Doctor Name:</label>
		<input type="text" name="doctorname" placeholder="Enter doctor name" required>

		<label>Invoice Number (Optional):</label>
		<input type="text" name="invoiceno" placeholder="Auto-generated if empty">

		<table id="itemsTable">
			<tr>
				<th>Service / Medicine Name</th>
				<th>Quantity</th>
				<th>Unit Price</th>
				<th>Total</th>
			</tr>
			<tr>
				<td><input type="text" name="name[]" placeholder="Test name / medicine" required></td>
				<td><input type="number" name="qty[]" min="1" value="1" oninput="calculateRow(this)" required></td>
				<td><input type="number" name="price[]" min="0" step="0.01" value="0" oninput="calculateRow(this)" required></td>
				<td><input type="number" name="total[]" value="0" readonly></td>
			</tr>
		</table>

		<button type="button" onclick="addRow()">+ Add Service</button><br>
		<button type="submit">Generate Medical Invoice</button>
	</form>
</div>

<script>
	function addRow() {
		const table = document.getElementById("itemsTable");
		const row = table.insertRow();
		row.innerHTML =
		'<td><input type="text" name="name[]" required></td>' +
		'<td><input type="number" name="qty[]" min="1" value="1" oninput="calculateRow(this)" required></td>' +
		'<td><input type="number" name="price[]" min="0" step="0.01" value="0" oninput="calculateRow(this)" required></td>' +
		'<td><input type="number" name="total[]" value="0" readonly></td>';
	}

	function calculateRow(input) {
		const row = input.parentElement.parentElement;
		const qty = parseFloat(row.cells[1].children[0].value) || 0;
		const price = parseFloat(row.cells[2].children[0].value) || 0;
		row.cells[3].children[0].value = (qty * price).toFixed(2);
	}

	document.getElementById("invoiceForm").addEventListener("submit", function() {
		let sum = 0;
		document.querySelectorAll('input[name="total[]"]').forEach(t => {
			sum += parseFloat(t.value) || 0;
		});
		const input = document.createElement("input");
		input.type = "hidden";
		input.name = "totalBill";
		input.value = sum.toFixed(2);
		this.appendChild(input);
	});
</script>

</body>
</html>
`))

	//---------------------------------------
	// INVOICE TEMPLATE
	//---------------------------------------
	invoiceTpl := template.Must(template.New("invoice").Parse(`
<!DOCTYPE html>
<html>
<head>
	<title>Medical Invoice</title>
	<style>
		body { font-family: 'Segoe UI'; background: #eaf4ff; padding: 20px; }
		.container { width: 80%; margin: auto; background: #fff; padding: 30px;
					 border-radius: 10px; box-shadow: 0 0 15px rgba(0,0,0,0.15); }
		h1, h3 { text-align: center; color: #1e4e79; }
		table { width: 100%; border-collapse: collapse; margin-top: 20px; }
		th, td { padding: 12px; border-bottom: 1px solid #ddd; }
		th { background-color: #1e88e5; color: white; }
		tr:nth-child(even) { background: #f4faff; }
		#total { font-weight: bold; font-size: 18px; background-color: #d3e8ff; }
	</style>
</head>
<body>

<div class="container">
	<h1>{{.StoreName}}</h1>
	<h3>Doctor: {{.DoctorName}}</h3>
	<h3>Patient: {{.UserName}}</h3>
	<h3>Invoice No: {{.InvoiceNo}} | Date: {{.Date}}</h3>

	<table>
		<tr>
			<th>Service / Medicine</th>
			<th>Qty</th>
			<th>Unit Price</th>
			<th>Total</th>
		</tr>
		{{range .Items}}
		<tr>
			<td>{{.Name}}</td>
			<td>{{.Qty}}</td>
			<td>{{printf "%.2f" .Price}}</td>
			<td>{{printf "%.2f" .Total}}</td>
		</tr>
		{{end}}
		<tr>
			<td colspan="3" id="total">Total Amount</td>
			<td id="total">{{printf "%.2f" .TotalBill}}</td>
		</tr>
	</table>
</div>

</body>
</html>
`))

	//---------------------------------------
	// ROUTES
	//---------------------------------------
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, nil)
	})

	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		names := r.Form["name[]"]
		qtys := r.Form["qty[]"]
		prices := r.Form["price[]"]

		var items []Item
		for i := range names {
			var qty int
			var price float64
			fmt.Sscan(qtys[i], &qty)
			fmt.Sscan(prices[i], &price)

			items = append(items, Item{
				Name:  names[i],
				Qty:   qty,
				Price: price,
				Total: float64(qty) * price,
			})
		}

		var totalBill float64
		fmt.Sscan(r.FormValue("totalBill"), &totalBill)

		invoiceNo := r.FormValue("invoiceno")
		if invoiceNo == "" {
			invoiceNo = fmt.Sprintf("%d", time.Now().Unix())
		}

		invoice := Invoice{
			StoreName:  "City Care Hospital",
			DoctorName: r.FormValue("doctorname"),
			UserName:   r.FormValue("username"),
			InvoiceNo:  invoiceNo,
			Date:       time.Now().Format("02-Jan-2006"),
			Items:      items,
			TotalBill:  totalBill,
		}

		invoiceTpl.Execute(w, invoice)
	})

	//---------------------------------------
	// SERVER START
	//---------------------------------------
	fmt.Println("Server running at port:", port)
	http.ListenAndServe(":"+port, nil)
}
