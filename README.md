products-services
==

## Build and run

To build:

```go build```

To run:

```./product-services```

To run Go tests:

```go test ./... -v```

```
% go test ./... -v
?   	github.com/matiasinsaurralde/product-services	[no test files]
=== RUN   TestHandler
=== RUN   TestHandler/list_directories
=== RUN   TestHandler/list_payments
=== RUN   TestHandler/get_payments
--- PASS: TestHandler (0.00s)
    --- PASS: TestHandler/list_directories (0.00s)
    --- PASS: TestHandler/list_payments (0.00s)
    --- PASS: TestHandler/get_payments (0.00s)
PASS
ok  	github.com/matiasinsaurralde/product-services/api	(cached)
=== RUN   TestNewWithBaseDir
=== RUN   TestNewWithBaseDir/initialize_with_nonexistent_directory
=== RUN   TestNewWithBaseDir/initialize_with_an_existing_directory
--- PASS: TestNewWithBaseDir (0.00s)
    --- PASS: TestNewWithBaseDir/initialize_with_nonexistent_directory (0.00s)
    --- PASS: TestNewWithBaseDir/initialize_with_an_existing_directory (0.00s)
=== RUN   TestParsePayments
--- PASS: TestParsePayments (0.00s)
=== RUN   TestListDirectories
=== RUN   TestListDirectories/list_single_valid_directory
=== RUN   TestListDirectories/list_invalid_directories
2022/07/18 03:08:00 invalid directory name 'xyz': parsing time "xyz" as "20060102": cannot parse "xyz" as "2006"
2022/07/18 03:08:00 invalid directory name '2022_07_17': parsing time "2022_07_17" as "20060102": cannot parse "_07_17" as "01"
2022/07/18 03:08:00 invalid directory name '2022': parsing time "2022" as "20060102": cannot parse "" as "01"
2022/07/18 03:08:00 invalid directory name '202207': parsing time "202207" as "20060102": cannot parse "" as "02"
2022/07/18 03:08:00 invalid directory name '202207177': parsing time "202207177": extra text: "7"
--- PASS: TestListDirectories (0.00s)
    --- PASS: TestListDirectories/list_single_valid_directory (0.00s)
    --- PASS: TestListDirectories/list_invalid_directories (0.00s)
=== RUN   TestListPayments
--- PASS: TestListPayments (0.00s)
=== RUN   TestGetPayments
--- PASS: TestGetPayments (0.00s)
=== RUN   TestValidateDirName
--- PASS: TestValidateDirName (0.00s)
=== RUN   TestValidateFileName
--- PASS: TestValidateFileName (0.00s)
PASS
ok  	github.com/matiasinsaurralde/product-services/payment	(cached)
```

## Test requests:

```
% curl http://localhost:9999/ ; echo
["20220717","20220718"]
% curl http://localhost:9999/20220717/ ; echo
["063000.payments","090000.payments"]
% curl http://localhost:9999/20220717/063000.payments ; echo
[{"asOf":20220717063000,"sequence":111,"amount":1000,"comment":"payment1"},{"asOf":20220717063000,"sequence":112,"amount":1500,"comment":"payment2"}]
% curl http://localhost:9999/20220717/111122223333.payments ; echo
not found
```
