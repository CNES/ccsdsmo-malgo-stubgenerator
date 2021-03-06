# CNES MO go Stub Generator

This generator creates stub and encoding code targeted to the MAL/go implementation of the CNES.
It follows an informal definition of a MAL language mapping for the go language.

The framework CCSDS\_MO\_StubGenerator from ESA is reused.

## Usage

### Command line

> java esa.mo.tools.stubgen.StubGenerator -t go -x \<xml dir> -r \<xml-ref dir> -R \<xsd dir> -o \<src output dir>

The classpath must include the following maven artifacts:
- int.esa.ccsds.mo	StubGenerator
- int.esa.ccsds.mo	StubGenerator\_Lib
- fr.cnes.mo		StubGenerator\_go

### Command options

The -t go option directs the generator to generate go language output. The link between the generic generator and the go generation package is dynamic, assuming that the StubGenerator_go artifact is in the classpath.

The -x \<xml dir> option defines the directory where the generator is supposed to find the input files to process. All xml files in that directory will be processed as CCSDS definitions according to the ServiceSchema model.

The -r \<xml-ref dir> defines the directory where the generator can find CCSDS service definitions required by the files to process. The directory should at least include the ServiceDefMAL.xml file. No code shall be generated for those files.

The -R \<xsd dir> defines the directory where the generator is supposed to find XML schemas for the XML files in the other directories.

The -o \<src output dir> defines the directory where the generator produces go code. The provided directory must have an "src" part which defines the origin of the go source tree. As go packages and the go source files tree are closely related, this option implicitely declares a base package for all the generated go source code.

## Generated code and language mapping

### area

An xml defined area is turned into a go package and its associated directory. The directory is created in the parent directory defined by the -o command line option. The directory name is the area name all in lower letters.

> Area|AreA -> \<src output dir>/area
>
> package \<base package>.area

### service

An xml defined service is turned into a go sub-package of the area package. The associated directory name is the service name all in lower letters.

> Service|SerVice -> \<src output dir>/area/service
>
> package \<base package>.area.service

Consumer stubs and provider stubs are also created in the service package. Those stubs are detailed below.

### types

An xml defined composite or enumeration type is turned into a go type in the package of the service where the type is defined in, or in the package of the area if it is an area level type. The name of the generated type is the xml type name with a capitalized first letter.  
Associated list types are generated together with the base type. They are given a "List" suffix to the base type name.

> compositetype|Compositetype -> Compositetype, CompositetypeList
>
> enumerationType -> EnumerationType -> EnumerationType, EnumerationTypeList

The generated go type \<type> is defined so that the associated *<type> type implements the mal.Element interface.

A Composite type field is generally declared as a go pointer if the field is nullable, and as a type value if it is not nullable.

Enumeration values are defined as simili constants, i.e. as variables with an upper letters name.

> \<enumX> \<value1> -> \<area|service>.\<ENUMX>\_\<VALUE1>

### constant values

Constant values such as area or service numbers are defined in their respective area or service package. They can be found in the helper.go files generated in their respective directories. They should be mostly hidden by the use of the generated stubs.

> \<area number> -> \<area>.AREA\_NUMBER
>
> \<service number> -> \<service>.SERVICE\_NUMBER
>
> \<operation number> -> \<service>.\<OPERATION>\_OPERATION\_NUMBER

### provider stubs

The malgo API defines a provider as a set of handlers registered in a context to answer to MAL operation calls. The generated provider stubs hide most of this linking code so that the programmer may focus on the actual provider logic.

The provider stubs of a service are defined directly in the service package. They are made of three entities: a provider interface, a provider structure, and a provider helper.

The provider interface is a go interface named \<service>.ProviderInterface. It defines a function for each operation defined in the service. The programmer supplies a provider to the mal as an object (go structure named MyProvider in the example below) implementing this interface. It is registered using a NewProvider generated function in the service.

The provider structure is a go structure named \<service>.Provider. It is created and returned by the call to \<service>.NewProvider. A mal.ClientContext is automatically created by the call, so that the function Close should eventually be called on this structure.

The provider helper is a go structure which is created and passed to the implementation each time an operation is called on the service. It encapsulates the mal.Transaction concept of the mal API. Helper structures are defined one for each operation of the service. The helper is typed and named by the operation, it provides functions related to the operation Interaction Pattern, with parameters as declared in the xml operation message.

	// My <service> implementation, must implement the interface ProviderInterface
	type MyProvider ...
	func (p *MyProvider) <Operation>(opHelper <service>.<Operation>Helper, ...) error {
		opHelper.Ack(...)
		...
		opHelper.Reply(...)
		return nil
	}
	
	// register the provider
	ctx, err := mal.NewContext(provider_url)
	defer ctx.Close()
	myProvider := MyProvider(...)
	provider, err := <service>.NewProvider(ctx, provider_name, &myProvider)
	defer provider.Close()

### consumer stubs

The malgo API provides a consumer with mal.Operation objects. An Operation object must be created (or reused) for each call to a service. The mal.Operation object is typed by the Interaction Pattern of the called service.  
The generated consumer stubs provides a similar interface, except that the generated operation objects are typed by the operation itself, enabling to hide all the coding/decoding of the parameters.

The consumer stubs of a service are defined directly in the service package. They are made of an initialization function Init, and of a set of operation structures, one for each operation of the service.

The operation specific structure is named \<service>.\<Operation>Operation. It is created by a call to the generated function \<service>.New\<Operation>Operation. It defines functions related to the Interaction Pattern of the operation, exactly as the original mal API does.

	ctx, err := mal.NewContext(consumer_url)
	defer ctx.Close()
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	defer cctx.Close()
	err = <service>.Init(cctx)
	
	operation, err := <service>.New<Operation>Operation(providerUri)
	ackValue, err := operation.Invoke(...)
	respValue, err := operation.GetResponse()

### stubs for a PubSub operation

A PubSub operation is a special case as it involves three actors instead of just a consumer and a provider. A more or less independant broker represents the provider side of the IP, while two consumers interact with the broker as a publisher and a subscriber. The malgo API provides two separate mal.Operation objects to help implement the consumer sides, and provides a generic implementation of the broker side in the mal.broker package. The generated consumer stubs for the PubSub operation cover both sides, and they are very similar to what is generated for the other IPs.

	ctx, err := mal.NewContext(consumer_url)
	defer ctx.Close()
	cctx, err := malapi.NewClientContext(ctx, "consumer")
	defer cctx.Close()
	err = <service>.Init(cctx)
	
	publishOp, err := <service>.New<Operation>PublisherOperation(brokerUri)
	err = publishOp.Register(<entityKey lists>)
	publishOp.Publish(updtHeaders, <updtVal type 1>, <updtVal type 2>)
	
	subscribeOp, err := <service>.New<Operation>SubscriberOperation(brokerUri)
	err = subscribeOp.Register(<subscriptions>)
	subId, updtHeaders, <updtVal type 1>, <updtVal type 2>, err := subscribeOp.GetNotify()

The generic broker of the mal.broker package may be deployed either as a standalone broker, or as a broker integrated with the publisher consumer. Examples of those two usages are provided in the tests of the mal.broker package. The generated provider stubs help to deploy a standalone broker by hiding most of the required processing. In that case a dedicated broker must be defined for each operation. All the brokers may use the same mal.Context.

	ctx, err := mal.NewContext(broker_url)
	defer ctx.Close()
	mybroker, err := <service>.New<Operation>Broker(ctx, "broker")
	defer mybroker.Close()
	brokerUri := mybroker.Uri()

The generated stubs for deploying a broker integrated with the publisher are integrated with the publisher consumer stubs. The only difference from a standard publisher consumer is that the operation constructor takes nil as an input broker URI. There is no other instruction which concerns the broker; everything is hidden in the publisher stubs. In that case the subscriber consumer must connect to the publisher consumer URI.

	subscribeOp, err := <service>.New<Operation>SubscriberOperation(nil)

The provider stubs use the generic broker implementation of the mal.broker package. If a specific behaviour needs to be implemented, use the malgo API instead.

### errors

The MAL specification defines errors as special messages in an Interaction Pattern with an error code and an optional extra information field whose type is variable. The type of the extra information field actually depends on the error code and on the operation which sends the error message. The generated stubs provides a limited help to hide the coding/decoding of this field.

The generated code uses the malapi.MalError type, which implements the Error interface of the standard go package error. Whenever a consumer call returns a not null error parameter, then this parameter may be checked to be a MalError. If it is the case, the client code may check the value of the Code field, and then cast the ExtraInformation field to the type declared in the service specification for this code and operation.

	ackValue, err := operation.Invoke(...)
	if err != nil {
		malErr, ok = err.(*malapi.MalError)
		if ok {
			errmsg, ok := malErr.ExtraInfo.(*mal.String)
		}
	}

In the provider implementation code, the error may be signaled by returning a MalError from the operation function.

	func (p *MyProvider) <Operation>(opHelper <service>.<Operation>Helper, ...) error {
		return malapi.NewMalError(<err code>, mal.NewString("err msg"))
	}

## Tests

### Basic tests

A small suite of tests is provided as a generator_test.go go file and a set of xml files. In order to execute the tests:

- process the xml files with the generator
- export GOPATH="\<path to the generated files>;\<path to malgo>;\<path to the local test file>"
- go test -v generator\_test.go

### Generating COM area

The content of the package github.com/CNES/ccsdsmo-malgo/com has been mainly generated by this generator from the standard XML definition file of the COM area. However the package includes additional files that could not be generated. So the content of the package directory can be safely overridden but not replaced by the output of the generator.

## Limitations

### Names

The go generator is not completely safe about name collision. It is good practice to use different names for all xml defined entities.

### Processing order

The xml input files are processed in lexical order. If there exist dependencies between files, then you should make sure the files are processed in the right order, or you will have to process the files in two separate passes.

### MAL area processing

The MAL area cannot be processed by the generator. Moreover the malgo implementation has not been changed to reflect the new go language mapping defined along with the generator. However a compatibility file named newmapping.go has been defined so that clients may use the new language mapping, including to the MAL area.

The generator makes use of this compatibility file and generates standard mapping code including for the MAL area.

### Publish errors

The MAL specification allows the publisher of a PubSub operation to issue PUBLISH\_ERROR messages. The generator does not provide a specific API to issue such messages, nor to process them.

An example of PUBLISH\_ERROR messages is provided in the publish\_error\_test.go file in the malgo implementation. It is based on the standard malgo API.

### Comments

The comments in the XML specification are not yet handled.
