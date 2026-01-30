# proto-type

Exploring functional, generic, and derived types in Protobuf/Grpc/Golang

# What is proto-type?

This is software intended to be used in the production of practical, non-mathematical/categorical/type-system adjacent software. So please forgive me as I'm (https://github.com/fredxfred) not a trained category theorist. My terminology, notation, rigor, and recognition/application of "basic"/typical results or constructions may be lacking; please let me know when it's bad enough to need fixing.

For type-theorists and the categorically enthused: create a **practical distributed, dynamically extensible, cross-language, cross-platform, fully-serializable, indexable/searchable/walkable, strongly-typed system for operating on categorical (ie functional, derived/generic) types of arbitrarily high dimension at scale, in the laziest way possible: applying the Yoneda lemma to a pre-existing type system with a rich set of natural transformations that closely models simplical sets, and lacks only a monoidal endofunctor on the category of types itself 😎.**

For everybody else: **strongly-typed, serializable, introspectable, functional/generic/derived types for distributed programming**, by making types runtime-operable/modifiable through functions capable of operating on other functions (functors, or higher-order functions) and on/to types themselves. Another way to think of it is:

**General higher order functions and derived types for APIs!**

## Technology Choice

Proto-type is intended to be a fully expressive type system for distributed systems and client-server interaction, with:

* Serializable (homoiconic), typed functions, and higher order functions

* Generics, Functors, Interfaces (ish), and other Derived Types

* A fully reflective type/object/functional system, even remotely

* By extension of the above, a *dynamically introspectable type system* that can be searched, walked, extended, or used to cast/transform objects across types without explicitly specifying an operation.

* The proper set of primitives, base constructions, and structure necessary to extend itself in a way that encode arbitrarily higher-order types/structures (or at least, the ones a "normal" user would care about)

Proto-type is based on Protobuf/gRPC/Golang because these three together already come very close to implementing most of the primitives necessary to bootstrap an extensible, arbitrarily derived, fully serializable, algebraic type system.

* [Protobuf](https://protobuf.dev/) is already a well-supported, cross-language, cross-platform, extensible, typed data format system meant for use in distributed systems. And [gRPC](https://grpc.io/) is a well-supported for implementing APIs (ie functions, or Services and Methods in gRPC lingo) as remote procedure calls over/with protobuf.

Very, very few serialization formats are as battle-tested, featureful, and adopted as these two (and even fewer still are typed: JSON+HTTP and Bytestreams+Files are probably the only two general data:transport pairs more widely adopted). gRPC has support for actual versioned and typed schemas with runtime bindings reflecting the actual constituent fields with actual typed primitives in a way that addresses many forward-backward-implementation-runtime compatibility problems that the JSON/HTTP ecosystem does not, as well as full HTTP/2 BIDI. And protobuf is actually strongly typed, not a bolt-on wrapper introducing types to a fundamentally untyped implementation like JSON. Also, there are good tools for automatically generating OpenAPI specs/equivalent HTTP server/client implementations of proto/grpc services via eg [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) (as well as [gRPC-Web](https://github.com/grpc/grpc-web)).

Google relies very heavily on protobuf+grpc across all their client libraries/code/APIs, so there are very good security advantages to building soemthing on *top* of protobuf and gRPC while modifying the internals/ecosystem as little as possible (or not at all): the parsing, serialization, platform, cryptographic, and network layer are not only already implemented but have just about the best maintenance and threat modelling/security stance you could ask for in a distributed communication protocol: probably better than any other in the world and certainly MUCH better than anything we could write ourselves, use from a less sophisticated software vendor, or extend from an existing github/academic project or programming language.

* Protobuf has support for a [runtime type reflection](https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/descriptor.proto), and gRPC for a dynamically/remotely invocable [reflection API](https://grpc.io/docs/guides/reflection/), as does [Go itself](https://go.dev/blog/laws-of-reflection) in its [language libraries](https://pkg.go.dev/reflect).

* Golang compiles to static binaries that can be serialized as bytes and sent to remote servers, or referenced in terms of their compiler version + build info + source code with a reasonable expectation that the same binary will be compiled, or invoked remotely with a reasonable expectation that linked libraries or runtime trickery substantially changes the semantics of the invoked code

Note that we do not need hard guarantees of full homoiconicity or reproducibility across build/remote rpcs/side effects/runtimes/trust boundaries. Yet :) For now let's just consider Go to have "mostly homoiconic binary executables" across "mostly similar execution runtimes" in a way that is good enough to consider there to be well-defined homomorphisms across eg building+executing, rpc send+recv, binary download+execution, etc. We will implement some of this

* Golang has a [strong base implementation for a fully reflective type system](https://pkg.go.dev/google.golang.org/protobuf/reflect/protoreflect) for proto/grpc

Google's DescriptorProto, FileDescriptorProto, and friends are very good starting points for extendng proto with more sophisticated dynamic/derived types based on the Proto and gRPC primitives and built-ins, which have a very rich set of "Natural Transformations" and compositions/equivalences across types already established.

Also, there is some prior work building in this area in eg [jhump's protobuilder](https://pkg.go.dev/github.com/jhump/protoreflect/v2/protobuilder) and related packages in his [protoreflect implementation](https://github.com/jhump/protoreflect). Although these don't go as far as defining or implementing derived/generic types in proto, they have most of the additional building blocks necessary (especially protoprint and protobuilder, with good ideas/pocs defined in other packages as well) for accomplishing this.

* Golang is strongly-typed and [treats functions as first-class citizens, with higher-order functions, function literals, and functions-as-data](https://go.dev/doc/codewalk/functions/) fully supported at runtime, with very good support for systems/network programming/use with protobuf and gRPC (and without being overly abstract/experiment/lacking in tooling)

Other programming languages have more sophisticated type systems, better functional programming support, more elegant syntax or whatever. Very few are also primarily statically linked and compiled to machine language, and able to operate on grpc services/network/systems as well as Go. This is crucial, because simple (de)serialization to/from proto and runtime types, is something we'll need to be doing a lot, as well as sending data to code.

* Protobuf/gRPC have strong, robust ecosystems beyond just Google itself

Companies like [Buf](https://buf.build/) are all-in on protobuf tooling and support itself, and provide their own strong useful ecosystem packages like [protovalidate](https://github.com/bufbuild/protovalidate) - many other large tech companies use Protobuf/gRPC either openly (eg Apple, Nvidia) or implicitly by contributing and using other software (like Envoy) that strongly leverages protobuf internally (eg for [Envoy's xDS](https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol) which will most likely eventually get tied in more directly to this project).

This is a major risk when considering similar project's like Kenton Varda's (Cloudflare employee) [Capn Proto](https://capnproto.org/) (which also has only a single working runtime implementation which invites multiple other kinds of risks/problems) and other gRPC/proto derivatives.

Although it's possible that at some point we may want to diverge from gRPC/proto implementation internals, or use slightly different tooling (eg an augmented protoc implementation), that introduces major maintenance/security/compatibility issues that we're better off avoiding as long as possible.

# Establishing a Sufficient Basis for Arbitrary Serializable N-Categorical Constructions

Proof sketch (using informal inline rpc message definitions for convenience):

```
Implement deferred evaluation over gRPC methods via dynamic dispatch, Dispatch({methodName: service.Method, request: myData})

Make Binaries fully reflective so that Dispatch({methodName: service.Method, request: myData, target?: targetService}) => Dispatch({binaryDescriptor: bdProto, methodName: service.Method, request: myData}) | Dispatch({dispatchCreds: creds, methodName: service.Method, request: myData}) | Dispatch({methodName: service.Method, request: myData}) to directly invoke a remote grpc service method on introspectable code-as-data if desired. This will also be useful locally for linking/loading other binaries at runtime in order to execute their gRPC services and workaround Go's limited plugin/dlopen support.

Allow for partial evaluation via
Dispatch({methodName: service.Method, request: someOtherMethod}) => Dispatch({methodName: service.Method, request: someOtherMethod})
Dispatch({Dispatch({methodName: service.Method, request: someOtherMethod}), request: target}) => messy notation, make it satisfy morphism requirements so that Dispatch becomes capable of representing Functors

Use reflection on FileDescriptorProto/DescriptorProto to support Functors on each so that entire services/objects can be lifted or used as the bases for derived types, which may itself be represented with

rpc DeriveType(FileDescriptorProto) returns(FileDescriptorProto);


Together with Disaptch this allows for the continuation-passing-style https://en.wikipedia.org/wiki/Continuation-passing_style construction of a monadic category around FileDescriptorProto (the catch-all protobuf construct for "types") https://en.wikipedia.org/wiki/Monad_(functional_programming)#Continuation_monad

Now we have an endofunctor defined over the category and, as we all know, a monad is a monoid in the category of endofunctors, so we can now easily represent any type transition we wish as a series of folded/accumulated (deferred Dispatch calls) chained upon on monad according to the "Builder" pattern. This allows us to represent proper higher-order-functions and generic/collections types, eg

Collection<T> is just 

rpc DeriveCollection(FileDescriptorProto /* this is T */) returns(FileDescriptorProto /* this is Collection<T> */);

and Tuple<A, ..> is just

rpc DeriveTuple({repeated FileDescriptorProto elems A, B, C) returns(FileDescriptorProto /* this is Tuple<A, B, C> */); 

with higher-order-functions being representable as fully sequence binarydescriptor protos/continuation passing into dynamically executed binarydescriptors.

Tada! We have a rich set of natural transformations https://en.wikipedia.org/wiki/Natural_transformation supported out of the box courtesy of golang/proto/grpc/reflection libraries and an obvious bijection between proto messages (DescriptorProto) and Types, and back to Types and other Types, as well as from a Type to a message (eg casting generic bytes as a type, or from one type to another), and representing both types as messages (FileDescriptorProto) and messages as types (DescriptorProto). So, we have an endofunctor not only from any Type-Type but from any Message->Type (the type is belongs to)->Type (any other type)->Message(representing this type) | Message(belonging to this type) (and similarly with functions). 

That's clearly a bicategory, and since the types' interior structures are runtime-introspectable via reflection APIs, we can now defer and partially evaluate structures of "chained" (Derived) and "paired" (partially evaluated deferred dispatch) morphisms between objects and/or types through a "cell" like structure as given by the S-expression structure: https://en.wikipedia.org/wiki/S-expression

Because this allows us to derive higher-order categorical structures across both the categories' objects and morphisms to an arbitrarily high degree of across either or both in conjunction (or, because proto messages are very clearly simplical https://en.wikipedia.org/wiki/Simplicial_set), we can see that this represents a "quasi-category" or rather the general ∞-category provided that the partially evaluable structures are formally encoded themselves as proper object/morphisms within our system.

```

TLDR: deferred dispatch, functions mapping Types->Types (and natural transformations for eg message->Types->message), partial evaluation (allowing Functor and Generic types to be represented) as properly-respected objects/functions within our type system is all you need to make proto a fully categorical type system, I think!
