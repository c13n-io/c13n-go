package itest

// testsCases is the list of tests to be executed. Tests will be
// executed in the same order as defined in testsCases array.
var testsCases = []*testCase{
	{
		name: "GetRoute",
		test: testGetRoute,
	},
	{
		name: "Subscribe Invoice Updates",
		test: testSubscribeInvoiceUpdates,
	},
	{
		name: "Send Payment",
		test: testSendPayment,
	},
	{
		name: "SendPayment with Subscribe Invoices",
		test: testSendPaymentSubscribeInvoiceUpdates,
	},
	{
		name: "SignMessage",
		test: testSignMessage,
	},
	{
		name: "LND Single Hop tests",
		test: testSingleHopTests,
	},
	{
		name: "LND Multi-hop tests",
		test: testMultiHopTests,
	},
	{
		name: "InitializeConnection",
		test: testInitializeConnection,
	},
	{
		name: "Test new lnchat",
		test: testNewLnchat,
	},
	{
		name: "GetSelfInfo",
		test: testGetSelfInfo,
	},
	{
		name: "ListNodes",
		test: testListNodes,
	},
	{
		name: "GetSelfBalance",
		test: testGetSelfBalance,
	},
	{
		name: "OpenChannel",
		test: testOpenChannel,
	},
	{
		name: "ConnectNode",
		test: testConnectNode,
	},
	{
		name: "DecodePayReq",
		test: testDecodePayReq,
	},
	{
		name: "CreateInvoice",
		test: testCreateInvoice,
	},
	{
		name: "LookupInvoice",
		test: testLookupInvoice,
	},
	{
		name: "SubscribePaymentUpdates",
		test: testSubscribePaymentUpdates,
	},
}
