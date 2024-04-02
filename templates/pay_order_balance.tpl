货币:{{.Order.Currency}}
主网:{{.Order.Network}}
金额:{{.Order.Price}}
{{ .OrderNoteTitle }}{{ .OrderNote }}

创建时间:{{TimestampToDatetime .Order.CreateTime}}
结束时间:{{TimestampToDatetime .Order.EndTime}}
