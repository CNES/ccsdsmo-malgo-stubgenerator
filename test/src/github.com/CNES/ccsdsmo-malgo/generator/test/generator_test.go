package test

import (
	"fmt"
	"testing"
	"time"
	
	"github.com/CNES/ccsdsmo-malgo/mal"
	malapi "github.com/CNES/ccsdsmo-malgo/mal/api"
	_ "github.com/CNES/ccsdsmo-malgo/mal/transport/tcp" // Needed to initialize TCP transport factory
	"github.com/CNES/ccsdsmo-malgo/mal/broker"
	"github.com/CNES/ccsdsmo-malgo/testsendarea/testsend1"
	"github.com/CNES/ccsdsmo-malgo/testsubmitarea"
	"github.com/CNES/ccsdsmo-malgo/testsubmitarea/testsubmit1"
	"github.com/CNES/ccsdsmo-malgo/testsubmitarea/testsubmit2"
	"github.com/CNES/ccsdsmo-malgo/testrequestarea"
	"github.com/CNES/ccsdsmo-malgo/testrequestarea/testrequest1"
	"github.com/CNES/ccsdsmo-malgo/testinvokearea/testinvoke1"
	"github.com/CNES/ccsdsmo-malgo/testprogressarea/testprogress1"
	"github.com/CNES/ccsdsmo-malgo/testprogressarea/testinout"
	"github.com/CNES/ccsdsmo-malgo/testpubsubarea"
	"github.com/CNES/ccsdsmo-malgo/testpubsubarea/testpubsub1"
)

const (
	provider_url   = "maltcp://127.0.0.1:16001"
	consumer_url   = "maltcp://127.0.0.1:16002"
	subscriber_url = "maltcp://127.0.0.1:16001"
	publisher_url  = "maltcp://127.0.0.1:16002"
	broker_url     = "maltcp://127.0.0.1:16003"
	provider_name  = "myprovider"
	broker_name  = "mybroker"
	varint = false
)

/*************************
 * implantation des providers
 *************************/

// Ce provider de test implante l'ensemble des services de test
type Provider string

// implantation du service testsend1:dosend
func (provider *Provider) Dosend(opHelper *testsend1.DosendHelper, value *mal.Integer) error {
	fmt.Printf("test.Provider(%s).dosend(%d)\n", *provider, *value)
	return nil
}

// implantation du service testsubmit1:dosubmit
func (provider *Provider) Dosubmit(opHelper *testsubmit1.DosubmitHelper, value *mal.Integer) error {
	fmt.Printf("test.Provider(%s).dosubmit(%d)\n", *provider, *value)
	if (*value) < 0 {
		return malapi.NewMalError(mal.ERROR_UNKNOWN, mal.NewString("Illegal input: negative value"))
	}
	return opHelper.Ack()
}

// implantation du service testsubmit2
func (provider *Provider) Dosubmit_a(opHelper *testsubmit2.Dosubmit_aHelper, value *mal.Integer) error {
	fmt.Printf("test.Provider(%s).dosubmit_a(%d)\n", *provider, *value)
	if (*value) < 0 {
		return malapi.NewMalError(testsubmitarea.ERROR_UNKNOWN_A, mal.NewString("Illegal input: negative value"))
	}
	return opHelper.Ack()
}
func (provider *Provider) Dosubmit_s(opHelper *testsubmit2.Dosubmit_sHelper, value *mal.Integer) error {
	fmt.Printf("test.Provider(%s).dosubmit_s(%d)\n", *provider, *value)
	if (*value) < 0 {
		return malapi.NewMalError(testsubmit2.ERROR_UNKNOWN_S, mal.NewString("Illegal input: negative value"))
	}
	return opHelper.Ack()
}
func (provider *Provider) Dosubmit_o(opHelper *testsubmit2.Dosubmit_oHelper, value *mal.Integer) error {
	fmt.Printf("test.Provider(%s).dosubmit_o(%d)\n", *provider, *value)
	if (*value) < 0 {
		return malapi.NewMalError(testsubmit2.DOSUBMIT_O_ERROR_UNKNOWN_O, mal.NewString("Illegal input: negative value"))
	}
	return opHelper.Ack()
}

// implantation du service testrequest1:dorequest
func (provider *Provider) Dorequest(opHelper *testrequest1.DorequestHelper, value *mal.Integer) error {
	fmt.Printf("test.Provider(%s).dorequest(%d) -> reply(%d)\n", *provider, *value, *value)
	return opHelper.Reply(value)
}

// implantation du service testinvoke1
func (provider *Provider) Doinvoke(opHelper *testinvoke1.DoinvokeHelper, value *mal.Integer) error {
	fmt.Printf("test.Provider(%s).dorequest(%d) -> ack(%d)\n", *provider, *value, *value)
	opHelper.Ack(value)
	outvalue := *value + 1
	fmt.Printf("test.Provider(%s).dorequest(%d) -> reply(%d)\n", *provider, *value, outvalue)
	return opHelper.Reply(&outvalue)
}

// implantation du service testprogress1
func (provider *Provider) Doprogress(opHelper *testprogress1.DoprogressHelper, baseIn *mal.Integer, rangeIn *mal.Integer) error {
	fmt.Printf("test.Provider(%s).doprogress(%d, %d) -> ack(%d)\n", *provider, *baseIn, *rangeIn, *baseIn)
	opHelper.Ack(baseIn)
	outValue := *baseIn
	for i := int32(*rangeIn)-1; i > 0; i-- {
		outValue ++
		fmt.Printf("test.Provider(%s).doprogress(%d, %d) -> update(%d)\n", *provider, *baseIn, *rangeIn, outValue)
		opHelper.Update(&outValue)
	}
	outValue ++
	fmt.Printf("test.Provider(%s).doprogress(%d, %d) -> reply(%d)\n", *provider, *baseIn, *rangeIn, outValue)
	return opHelper.Reply(&outValue)
}

// implantation du service TestInout
func (provider *Provider) ListEnum(opHelper *testinout.ListEnumHelper, list *testrequestarea.InOutList) error {
	listSize := 0
	if list != nil {
		listSize = list.Size()
	}
	fmt.Printf("test.Provider(%s).ListEnum(%v) -> ack(%d)\n", *provider, *list, listSize)
	opHelper.Ack(mal.NewInteger(int32(listSize)))
	for i := 0; i < listSize; i++ {
		outValue := (*list)[i]
		fmt.Printf("test.Provider(%s).ListEnum -> update(%v)\n", *provider, *outValue)
		opHelper.Update(outValue)
	}
	fmt.Printf("test.Provider(%s).doprogress -> reply\n", *provider)
	return opHelper.Reply()
}

// broker implementation

type IntegerUpdateValueHandler struct {
	list   *mal.IntegerList
	values mal.IntegerList
}

func NewIntegerUpdateValueHandler() *IntegerUpdateValueHandler {
	return new(IntegerUpdateValueHandler)
}

func (handler *IntegerUpdateValueHandler) DecodeUpdateValueList(body mal.Body) error {
	p, err := body.DecodeLastParameter(mal.NullIntegerList, false)
	if err != nil {
		return err
	}
	list := p.(*mal.IntegerList)
	listSize := 0
	var listInts []*mal.Integer
	if list != nil {
		listSize = len([]*mal.Integer(*list))
		listInts = []*mal.Integer(*list)
	}
	fmt.Printf("Broker.Publish, DecodeUpdateValueList -> %d [", listSize)
	for i := 0; i < listSize; i++ {
		fmt.Printf("%v ", *listInts[i])
	}
	fmt.Printf("]\n")

	handler.list = list
	handler.values = mal.IntegerList(make([]*mal.Integer, 0, handler.list.Size()))

	return nil
}

func (handler *IntegerUpdateValueHandler) UpdateValueListSize() int {
	return handler.list.Size()
}

func (handler *IntegerUpdateValueHandler) AppendValue(idx int) {
	handler.values = append(handler.values, ([]*mal.Integer)(*handler.list)[idx])
}

func (handler *IntegerUpdateValueHandler) EncodeUpdateValueList(body mal.Body) error {
	err := body.EncodeLastParameter(&handler.values, false)
	if err != nil {
		return err
	}
	handler.values = handler.values[:0]
	return nil
}

func (handler *IntegerUpdateValueHandler) ResetValues() {
	handler.values = handler.values[:0]
}

// implantation du service testrequest1:testEnum
func (provider *Provider) TestEnum(opHelper *testrequest1.TestEnumHelper, value *testrequestarea.InOut) error {
	outValue := testrequestarea.INOUT_OUT
	fmt.Printf("test.Provider(%s).testEnum(%d) -> reply(%d)\n", *provider, *value, outValue)
	return opHelper.Reply(&outValue)
}

// Test functions

func TestSend(t *testing.T) {
	// Waits socket closing from previous test
	time.Sleep(250 * time.Millisecond)
	
	ctx, err := mal.NewContext(provider_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer ctx.Close()
	
	providerImpl := Provider(provider_name)
	providerUri := mal.NewURI(provider_url + "/" + provider_name)
	provider, err := testsend1.NewProvider(ctx, provider_name, &providerImpl)
	if err != nil {
		t.Fatal("Error creating testsend1 provider, ", err)
		return
	}
	defer provider.Close()
	
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	if err != nil {
		t.Fatal("Error creating consumer context, ", err)
		return
	}
	err = testsend1.Init(cctx)
	if err != nil {
		t.Fatal("Error initializing testsend1 consumer, ", err)
		return
	}
	operation, err := testsend1.NewDosendOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating dosend operation, ", err)
		return
	}
	err = operation.Send(mal.NewInteger(5))
	if err != nil {
		t.Fatal("Error calling dosend operation, ", err)
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func TestSubmit1(t *testing.T) {
	// Waits socket closing from previous test
	time.Sleep(250 * time.Millisecond)
	
	ctx, err := mal.NewContext(provider_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer ctx.Close()
	
	providerImpl := Provider(provider_name)
	providerUri := mal.NewURI(provider_url + "/" + provider_name)
	provider, err := testsubmit1.NewProvider(ctx, provider_name, &providerImpl)
	if err != nil {
		t.Fatal("Error creating testsubmit1 provider, ", err)
		return
	}
	defer provider.Close()
	
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	if err != nil {
		t.Fatal("Error creating consumer context, ", err)
		return
	}
	err = testsubmit1.Init(cctx)
	if err != nil {
		t.Fatal("Error initializing testsubmit1 consumer, ", err)
		return
	}
	operation, err := testsubmit1.NewDosubmitOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating dosubmit operation, ", err)
		return
	}
	err = operation.Submit(mal.NewInteger(5))
	if err != nil {
		t.Fatal("Error calling dosubmit operation, ", err)
		return
	}
	operation, err = testsubmit1.NewDosubmitOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating dosubmit operation, ", err)
		return
	}
	err = operation.Submit(mal.NewInteger(-1))
	if err == nil {
		t.Fatal("Calling dosubmit operation returns no error")
		return
	}
	malErr, ok := err.(*malapi.MalError)
	if !ok {
		t.Fatal("Error returned by dosubmit operation is no MalError")
		return
	}
	if malErr.Code != mal.ERROR_UNKNOWN {
		t.Fatal("Unexpected error code")
		return
	}
	errmsg, ok := malErr.ExtraInfo.(*mal.String)
	if !ok {
		t.Fatal("ExtraInfo returned by dosubmit operation is no String")
		return
	}
	t.Logf("dosubmit error %v: %v\n", malErr.Code, *errmsg)
	time.Sleep(500 * time.Millisecond)
}

func TestSubmit2(t *testing.T) {
	// Waits socket closing from previous test
	time.Sleep(250 * time.Millisecond)
	
	ctx, err := mal.NewContext(provider_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer ctx.Close()
	
	providerImpl := Provider(provider_name)
	providerUri := mal.NewURI(provider_url + "/" + provider_name)
	provider, err := testsubmit2.NewProvider(ctx, provider_name, &providerImpl)
	if err != nil {
		t.Fatal("Error creating testsubmit2 provider, ", err)
		return
	}
	defer provider.Close()
	
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	if err != nil {
		t.Fatal("Error creating consumer context, ", err)
		return
	}
	err = testsubmit2.Init(cctx)
	if err != nil {
		t.Fatal("Error initializing testsubmit1 consumer, ", err)
		return
	}
	
	operationA, err := testsubmit2.NewDosubmit_aOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating dosubmit_a operation, ", err)
		return
	}
	err = operationA.Submit(mal.NewInteger(5))
	if err != nil {
		t.Fatal("Error calling dosubmit_a operation, ", err)
		return
	}
	operationA, err = testsubmit2.NewDosubmit_aOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating dosubmit_a operation, ", err)
		return
	}
	err = operationA.Submit(mal.NewInteger(-1))
	if err == nil {
		t.Fatal("Calling dosubmit_a operation returns no error")
		return
	}
	malErr, ok := err.(*malapi.MalError)
	if !ok {
		t.Fatal("Error returned by dosubmit_a operation is no MalError")
		return
	}
	if malErr.Code != testsubmitarea.ERROR_UNKNOWN_A {
		t.Fatal("Unexpected error code")
		return
	}
	errmsg, ok := malErr.ExtraInfo.(*mal.String)
	if !ok {
		t.Fatal("ExtraInfo returned by dosubmit_a operation is no String")
		return
	}
	t.Logf("dosubmit_a error %v: %v\n", malErr.Code, *errmsg)
	
	operationS, err := testsubmit2.NewDosubmit_sOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating dosubmit_s operation, ", err)
		return
	}
	err = operationS.Submit(mal.NewInteger(5))
	if err != nil {
		t.Fatal("Error calling dosubmit_s operation, ", err)
		return
	}
	operationS, err = testsubmit2.NewDosubmit_sOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating dosubmit_s operation, ", err)
		return
	}
	err = operationS.Submit(mal.NewInteger(-1))
	if err == nil {
		t.Fatal("Calling dosubmit_s operation returns no error")
		return
	}
	malErr, ok = err.(*malapi.MalError)
	if !ok {
		t.Fatal("Error returned by dosubmit_s operation is no MalError")
		return
	}
	if malErr.Code != testsubmit2.ERROR_UNKNOWN_S {
		t.Fatal("Unexpected error code")
		return
	}
	errmsg, ok = malErr.ExtraInfo.(*mal.String)
	if !ok {
		t.Fatal("ExtraInfo returned by dosubmit_s operation is no String")
		return
	}
	t.Logf("dosubmit_s error %v: %v\n", malErr.Code, *errmsg)
	
	operationO, err := testsubmit2.NewDosubmit_oOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating dosubmit_o operation, ", err)
		return
	}
	err = operationO.Submit(mal.NewInteger(5))
	if err != nil {
		t.Fatal("Error calling dosubmit_o operation, ", err)
		return
	}
	operationO, err = testsubmit2.NewDosubmit_oOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating dosubmit_o operation, ", err)
		return
	}
	err = operationO.Submit(mal.NewInteger(-1))
	if err == nil {
		t.Fatal("Calling dosubmit_o operation returns no error")
		return
	}
	malErr, ok = err.(*malapi.MalError)
	if !ok {
		t.Fatal("Error returned by dosubmit_o operation is no MalError")
		return
	}
	if malErr.Code != testsubmit2.DOSUBMIT_O_ERROR_UNKNOWN_O {
		t.Fatal("Unexpected error code")
		return
	}
	errmsg, ok = malErr.ExtraInfo.(*mal.String)
	if !ok {
		t.Fatal("ExtraInfo returned by dosubmit_o operation is no String")
		return
	}
	t.Logf("dosubmit_o error %v: %v\n", malErr.Code, *errmsg)

	time.Sleep(500 * time.Millisecond)
}

func TestRequest(t *testing.T) {
	// Waits socket closing from previous test
	time.Sleep(250 * time.Millisecond)
	
	ctx, err := mal.NewContext(provider_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer ctx.Close()
	
	providerImpl := Provider(provider_name)
	providerUri := mal.NewURI(provider_url + "/" + provider_name)
	provider, err := testrequest1.NewProvider(ctx, provider_name, &providerImpl)
	if err != nil {
		t.Fatal("Error creating testrequest1 provider, ", err)
		return
	}
	defer provider.Close()
	
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	if err != nil {
		t.Fatal("Error creating consumer context, ", err)
		return
	}
	err = testrequest1.Init(cctx)
	if err != nil {
		t.Fatal("Error initializing testrequest1 consumer, ", err)
		return
	}
	operation, err := testrequest1.NewDorequestOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating dorequest operation, ", err)
		return
	}
	inValue := mal.NewInteger(5)
	outValue, err := operation.Request(inValue)
	if err != nil {
		t.Fatal("Error calling dorequest operation, ", err)
		return
	}
	if outValue == nil || *outValue != *inValue {
		t.Fatal("Bad return value from dorequest operation")
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func TestInvoke(t *testing.T) {
	// Waits socket closing from previous test
	time.Sleep(250 * time.Millisecond)
	
	ctx, err := mal.NewContext(provider_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer ctx.Close()
	
	providerImpl := Provider(provider_name)
	providerUri := mal.NewURI(provider_url + "/" + provider_name)
	provider, err := testinvoke1.NewProvider(ctx, provider_name, &providerImpl)
	if err != nil {
		t.Fatal("Error creating testinvoke1 provider, ", err)
		return
	}
	defer provider.Close()
	
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	if err != nil {
		t.Fatal("Error creating consumer context, ", err)
		return
	}
	err = testinvoke1.Init(cctx)
	if err != nil {
		t.Fatal("Error initializing testinvoke1 consumer, ", err)
		return
	}
	operation, err := testinvoke1.NewDoinvokeOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating doinvoke operation, ", err)
		return
	}
	inValue := mal.NewInteger(5)
	outValue, err := operation.Invoke(inValue)
	if err != nil {
		t.Fatal("Error calling doinvoke operation, ", err)
		return
	}
	if outValue == nil || *outValue != *inValue {
		t.Fatal("Bad ack value from doinvoke operation")
		return
	}
	outValue, err = operation.GetResponse()
	if err != nil {
		t.Fatal("Error calling GetResponse operation, ", err)
		return
	}
	if outValue == nil || (*outValue - *inValue) != 1 {
		t.Fatal("Bad return value from doinvoke operation")
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func TestProgress(t *testing.T) {
	// Waits socket closing from previous test
	time.Sleep(250 * time.Millisecond)
	
	ctx, err := mal.NewContext(provider_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer ctx.Close()
	
	providerImpl := Provider(provider_name)
	providerUri := mal.NewURI(provider_url + "/" + provider_name)
	provider, err := testprogress1.NewProvider(ctx, provider_name, &providerImpl)
	if err != nil {
		t.Fatal("Error creating testprogress1 provider, ", err)
		return
	}
	defer provider.Close()
	
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	if err != nil {
		t.Fatal("Error creating consumer context, ", err)
		return
	}
	err = testprogress1.Init(cctx)
	if err != nil {
		t.Fatal("Error initializing testprogress1 consumer, ", err)
		return
	}
	operation, err := testprogress1.NewDoprogressOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating doprogress operation, ", err)
		return
	}
	inbase := mal.NewInteger(5)
	inrange := mal.NewInteger(4)
	outValue, err := operation.Progress(inbase, inrange)
	if err != nil {
		t.Fatal("Error calling doprogress operation, ", err)
		return
	}
	if outValue == nil || *outValue != *inbase {
		t.Fatal("Bad ack value from doprogress operation")
		return
	}
	
	next := *inbase + 1
	for endUpdates := false; !endUpdates; {
		outValue, err = operation.GetUpdate()
		if err != nil {
			t.Fatal("Error calling GetUpdate operation, ", err)
			return
		}
		if outValue == nil {
			endUpdates = true
		} else if *outValue == next {
			next = next+1
		} else {
			t.Fatal("Bad update value from doprogress operation")
			return
		}
	}
	outValue, err = operation.GetResponse()
	if err != nil {
		t.Fatal("Error calling GetResponse operation, ", err)
		return
	}
	if (outValue == nil) || (*outValue != next) || *outValue != (*inbase+*inrange) {
		t.Fatal("Bad return value from doprogress operation")
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func TestPubsub(t *testing.T) {
	// Waits socket closing from previous test
	time.Sleep(250 * time.Millisecond)
	
	brokerCtx, err := mal.NewContext(broker_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer brokerCtx.Close()
	
	mybroker, err := testpubsub1.NewDopubsubBroker(ctx, "broker")
	if err != nil {
		t.Fatal("Error creating broker, ", err)
		return
	}
	defer mybroker.Close()

	brokerUri := mybroker.Uri()
	
	ctx, err := mal.NewContext(consumer_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer ctx.Close()
	
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	if err != nil {
		t.Fatal("Error creating consumer context, ", err)
		return
	}
	defer cctx.Close()
	err = testpubsub1.Init(cctx)
	if err != nil {
		t.Fatal("Error initializing testpubsub1 consumer, ", err)
		return
	}
	cctx.SetDomain(mal.IdentifierList([]*mal.Identifier{mal.NewIdentifier("spacecraft1"), mal.NewIdentifier("payload"), mal.NewIdentifier("camera")}))
	publishOp, err := testpubsub1.NewDopubsubPublisherOperation(brokerUri)
	if err != nil {
		t.Fatal("Error creating dopubsub publish operation, ", err)
		return
	}
	ekpub1 := &mal.EntityKey{mal.NewIdentifier("key1"), mal.NewLong(1), mal.NewLong(1), mal.NewLong(1)}
	var eklist = mal.EntityKeyList([]*mal.EntityKey{ekpub1})
	err = publishOp.Register(&eklist)
	if err != nil {
		t.Fatal("Error registering publisher operation, ", err)
		return
	}
	
	subscribeOp, err := testpubsub1.NewDopubsubSubscriberOperation(brokerUri)
	if err != nil {
		t.Fatal("Error creating dopubsub subscribe operation, ", err)
		return
	}
	domains := mal.IdentifierList([]*mal.Identifier{mal.NewIdentifier("*")})
	eksub := &mal.EntityKey{mal.NewIdentifier("key1"), mal.NewLong(0), mal.NewLong(0), mal.NewLong(0)}
	var erlist = mal.EntityRequestList([]*mal.EntityRequest{
		&mal.EntityRequest{
			&domains, true, true, true, true, mal.EntityKeyList([]*mal.EntityKey{eksub}),
		},
	})
	var subid = mal.Identifier("MySubscription")
	subs := &mal.Subscription{subid, erlist}
	err = subscribeOp.Register(subs)
	if err != nil {
		t.Fatal("Error registering subscriber operation, ", err)
		return
	}
	
	inbase := mal.NewInteger(0)
	next := inbase
	for i := 0; i < 4; i++ {
		updthdr1 := &mal.UpdateHeader{*mal.TimeNow(), *cctx.Uri, mal.UPDATETYPE_CREATION, *ekpub1}
		updtHdrlist1 := mal.UpdateHeaderList([]*mal.UpdateHeader{updthdr1})
		updt1 := next
		updtlist1 := mal.IntegerList([]*mal.Integer{updt1})
		publishOp.Publish(&updtHdrlist1, &updtlist1)
		next = mal.NewInteger(int32(*next)+1)
	}
	
	next = inbase
	for i := 0; i < 4; i++ {
		_, _, outUpdt, err := subscribeOp.GetNotify()
		if err != nil {
			t.Fatal("Error calling dopubsub notify operation, ", err)
			return
		}
		if outUpdt == nil || outUpdt.Size() != 1 || *(*outUpdt)[0] != *next {
			t.Fatal("Bad update value from dopubsub notify operation")
			return
		}
		next = mal.NewInteger(int32(*next)+1)
	}
	
	time.Sleep(500 * time.Millisecond)
}

func TestPubsubLocal(t *testing.T) {
	// Waits socket closing from previous test
	time.Sleep(250 * time.Millisecond)
	
	ctx, err := mal.NewContext(consumer_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer ctx.Close()
	
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	if err != nil {
		t.Fatal("Error creating consumer context, ", err)
		return
	}
	defer cctx.Close()
	cctx.SetDomain(mal.IdentifierList([]*mal.Identifier{mal.NewIdentifier("spacecraft1"), mal.NewIdentifier("payload"), mal.NewIdentifier("camera")}))
	
	err = testpubsub1.Init(cctx)
	if err != nil {
		t.Fatal("Error initializing testpubsub1 consumer, ", err)
		return
	}
	publishOp, err := testpubsub1.NewDopubsubPublisherOperation(nil)
	if err != nil {
		t.Fatal("Error creating dopubsub publish operation, ", err)
		return
	}
	ekpub1 := &mal.EntityKey{mal.NewIdentifier("key1"), mal.NewLong(1), mal.NewLong(1), mal.NewLong(1)}
	var eklist = mal.EntityKeyList([]*mal.EntityKey{ekpub1})
	err = publishOp.Register(&eklist)
	if err != nil {
		t.Fatal("Error registering publisher operation, ", err)
		return
	}
	brokerUri := cctx.Uri
	
	subscribeOp, err := testpubsub1.NewDopubsubSubscriberOperation(brokerUri)
	if err != nil {
		t.Fatal("Error creating dopubsub subscribe operation, ", err)
		return
	}
	domains := mal.IdentifierList([]*mal.Identifier{mal.NewIdentifier("*")})
	eksub := &mal.EntityKey{mal.NewIdentifier("key1"), mal.NewLong(0), mal.NewLong(0), mal.NewLong(0)}
	var erlist = mal.EntityRequestList([]*mal.EntityRequest{
		&mal.EntityRequest{
			&domains, true, true, true, true, mal.EntityKeyList([]*mal.EntityKey{eksub}),
		},
	})
	var subid = mal.Identifier("MySubscription")
	subs := &mal.Subscription{subid, erlist}
	err = subscribeOp.Register(subs)
	if err != nil {
		t.Fatal("Error registering subscriber operation, ", err)
		return
	}
	
	inbase := mal.NewInteger(0)
	next := inbase
	for i := 0; i < 4; i++ {
		updthdr1 := &mal.UpdateHeader{*mal.TimeNow(), *cctx.Uri, mal.UPDATETYPE_CREATION, *ekpub1}
		updtHdrlist1 := mal.UpdateHeaderList([]*mal.UpdateHeader{updthdr1})
		updt1 := next
		updtlist1 := mal.IntegerList([]*mal.Integer{updt1})
		publishOp.Publish(&updtHdrlist1, &updtlist1)
		next = mal.NewInteger(int32(*next)+1)
	}
	
	next = inbase
	for i := 0; i < 4; i++ {
		_, _, outUpdt, err := subscribeOp.GetNotify()
		if err != nil {
			t.Fatal("Error calling dopubsub notify operation, ", err)
			return
		}
		if outUpdt == nil || outUpdt.Size() != 1 || *(*outUpdt)[0] != *next {
			t.Fatal("Bad update value from dopubsub notify operation")
			return
		}
		next = mal.NewInteger(int32(*next)+1)
	}
	
	time.Sleep(500 * time.Millisecond)
}

func TestRequestInout(t *testing.T) {
	// Waits socket closing from previous test
	time.Sleep(250 * time.Millisecond)
	
	ctx, err := mal.NewContext(provider_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer ctx.Close()
	
	providerImpl := Provider(provider_name)
	providerUri := mal.NewURI(provider_url + "/" + provider_name)
	provider, err := testrequest1.NewProvider(ctx, provider_name, &providerImpl)
	if err != nil {
		t.Fatal("Error creating testrequest1 provider, ", err)
		return
	}
	defer provider.Close()
	
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	if err != nil {
		t.Fatal("Error creating consumer context, ", err)
		return
	}
	err = testrequest1.Init(cctx)
	if err != nil {
		t.Fatal("Error initializing testrequest1 consumer, ", err)
		return
	}
	operation, err := testrequest1.NewTestEnumOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating testEnum operation, ", err)
		return
	}
	outValue, err := operation.Request(&testrequestarea.INOUT_IN)
	if err != nil {
		t.Fatal("Error calling testEnum operation, ", err)
		return
	}
	if outValue == nil || *outValue != testrequestarea.INOUT_OUT {
		t.Fatal("Bad return value from testEnum operation")
		return
	}
	time.Sleep(500 * time.Millisecond)
}

func TestInoutListEnum(t *testing.T) {
	// Waits socket closing from previous test
	time.Sleep(250 * time.Millisecond)
	
	ctx, err := mal.NewContext(provider_url)
	if err != nil {
		t.Fatal("Error creating context, ", err)
		return
	}
	defer ctx.Close()
	
	providerImpl := Provider(provider_name)
	providerUri := mal.NewURI(provider_url + "/" + provider_name)
	provider, err := testinout.NewProvider(ctx, provider_name, &providerImpl)
	if err != nil {
		t.Fatal("Error creating testrequest1 provider, ", err)
		return
	}
	defer provider.Close()
	
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	if err != nil {
		t.Fatal("Error creating consumer context, ", err)
		return
	}
	err = testinout.Init(cctx)
	if err != nil {
		t.Fatal("Error initializing testrequest1 consumer, ", err)
		return
	}
	operation, err := testinout.NewListEnumOperation(providerUri)
	if err != nil {
		t.Fatal("Error creating ListEnum operation, ", err)
		return
	}
	listInout := testrequestarea.NewInOutList(2)
	(*listInout)[0] = &testrequestarea.INOUT_IN
	(*listInout)[1] = &testrequestarea.INOUT_OUT
	ackValue, err := operation.Progress(listInout)
	if err != nil {
		t.Fatal("Error calling ListEnum operation, ", err)
		return
	}
	if ackValue == nil || int(*ackValue) != 2 {
		t.Fatal("Bad return value from ListEnum operation")
		return
	}
	for i := 0; i < listInout.Size(); i++ {
		outValue, err := operation.GetUpdate()
		if err != nil {
			t.Fatal("Error calling GetUpdate operation, ", err)
			return
		}
		if outValue == nil || *outValue != *(*listInout)[i] {
			t.Fatal("Bad update value from ListEnum operation")
			return
		}
	}
	err = operation.GetResponse()
	if err != nil {
		t.Fatal("Error calling GetResponse operation, ", err)
		return
	}
	time.Sleep(500 * time.Millisecond)
}

