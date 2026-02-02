# Projections <> Proto-Type

Suppose I have a protobuf message type Foo, that I would like to convert to a different protobuf message type. There are, roughly, three "conversions" to consider in implementing this:

1. The type conversion from Foo to the new type Foo'. There are many ways we could define or represent these "type conversions". The general class of type conversions can be given simply as F: Type -> Type

2. The instance conversion from a message of type Foo to type Foo'. This clearly has a level of "dependency" on the choice of F and the structure of both F and Foo'.

3. The communication/representation of this conversion in a way that can be understood across a distributed system with a common type registry/resolution mechanism.

# Projections

A naive, maximmally general approach would be to simply create an entirely new .proto file defining the Foo' type, with a Converter service with a Convert method that can convert messages of type Foo to Foo', and code for implementing that conversion.

However, there is a problem with the naive approach: it models F as merely Foo->Foo' rather than the more general Type->Type. In some cases, we may only want to represent this "conversion" as something defined on Foo: f: {Foo} -> {Foo'}, where {Foo} is a singleton set consisting only of Foo, likewise for {Foo'}.

But there are many transformations that be represented in a way that is more generic. For example, the conversion logic is expressed in terms of field transformations; f constructs Foo' by first having it "inherit" some fields from Foo, then specify new fields of its own. Or perhaps instead of being expressed in terms of specific field transformations, a conversion g is defined purely in terms of operations on generic fields, eg:

```

message Foo {
  string name = 1;
  int id = 2;
  string desc = 3;
  int num_replicas = 4;
}

// f: extract fields name, desc, then add new boolean field "original_type"

message Foo' {
  string name = 1;
  string desc = 2;
  bool original_type = 3;
}

// g: present string fields, rename int fields to {{name}}_original, add new timestamp field "snapshot_time",

message Foo'' {
  string name = 1;
  int id_original = 2;
  string desc = 3;
  int num_replies_original = 4;
  Timestamp snapshot_time = 5;
}

```

As we know, F:Type->Type within the protobuf type system can be given by F: DescriptorProto -> DescriptorProto. However, in the example above, f is defined in a way with unclear semantics if the name and desc field are not defined in the original type: we may want it to be actually undefined on messages without a name or description field. Or, we may want to define our f only on type Foo. Our example g is well-defined for any DescriptorProto. So, although F: Type->Type, since not every f is well-defined for all Types, we really have F:Poset(Type)->Type | Undefined, or need some kind of F' for filtering/establishing inclusion or exclusion criteria across the set of all types, which would be F': DescriptorProto->boolean

Also, clearly, there is a class of conversions in which the type conversion and the instance conversion can both be represented the same way: f as defined above is also sufficient for defining the implementation of Foo foo -> Foo' foo', as is g if snapshot_time can be left as its "default"/unset value, or if we had a simple way to provide default evaluations inline to the conversion definition. However, **not all transformations are necessarily able to be represented so compactly**: for example, to convert from ContainerInstance to DNSZone we might have to look up where the Container is running against some external state/dynamically, if the DNS Zone information is not directly encoded within the ContainerInstance message anywhere. Or, perhaps the DNSZone message requires setting some COMPUTE_PROVIDER enum that we need to define in terms of some kind of complex evaluation on ContainerInstances, or perhaps we need snapshot_time to represent the actual conversion time (and so define it in terms of some evaluation/logic independent of the input type).

The first set could clearly be defined in terms of FieldDescriptorProto. The second would also need a way to represent default values, and the third, evaluation logic/"providers" for dynamically evaluating certain fields' values. If those evaluations could be parameterized in terms of the input message, that could also be expressed as an evaluation.

```

message SimpleProjection {
  message FieldProjection {
    FieldDescriptorProto source = 1;
    FieldDescriptorProto target = 2;
  }
  repeated FieldProjection field_projections = 3;
}

message BasisProjection {
  message FieldProjection {
    oneof Source {
      FieldDescriptorProto field_source = 1;
      google.protobuf.Any value_source = 2;
    }
    FieldDescriptorProto target = 3
  }
  repeated FieldProjection field_projections = 3;
}

message DynamicProjection {
  message FieldProjection {
    oneof Source {
      FieldDescriptorProto field_source = 1;
      google.protobuf.Any value_source = 2;
      // For some kind of generic "provider" taking either google.protobuf.Empty or the input type, FieldDescriptorProto, etc,
      MethodDescriptorProto method_source = 3;
    }
    FieldDescriptorProto target = 3
  }
  repeated FieldProjection field_projections = 3;
}

// TODO(tweak this so that RecursiveProjection truly recurses)
message RecursiveProjection {
  message FieldProjection {
    repeated field_sources = 1;
    // Empty = identity projection
    MethodDescriptorProto method = 2;
    FieldDescriptorProto target = 3;
  }
  repeated FieldProjection field_projections = 3;
}
```

These types of conversions, which we may call "projections", have convenient properties: they provide the full set of necessary information for evaluating

* Foo' as a function of Foo, evaluated as a lazily derived type without upfront creation of a new .proto, implementation, conversion service
* Foo:foo->Foo'(foo') without a language-specific implementing of the logic for the conversion, since it is merely expressed in terms of fields and method evaluations on fields
* Whether or not Foo has a well-defined Foo' for the specific projection (whether it possesses fields matching the field_sources).

and are expressed in terms of the existing reflective DescriptorProto types. So for a given projection P, we have P: Type-> Type | undefined and P: x of Type X -> y of Type Y | undefined, in a way where we can also reason about the common structures of valid X and generated Y. These properties are categorically/algebraicaly very nice!

## Field Representations

Protobuf tends to suggest the use of field masks for these kinds of "projections". However, these are fundamentally different from our construction: they are moreso ways of "extracting" information from types rather than *converting* between types. Additionally, they are defined in terms of field names, and incapable of representing operations on repeated field types and truly dynamic logic.

This is an inconsistent representation of protobuf fields (using their field name in client-side code, in a way that is meant to be passed to the service, where the field could have been renamed or not yet defined) that doesn't rely on their actual field_numbers, so it's a pretty big mistake to build on top of this fundamentally leaky abstraction. However, FieldDescriptorProto is a bit clunky because it also involves specifying both field names and numbers. We could use a much more compact/efficient representation with less boilerplate

```proto

message RecursiveProjection {
  message FieldProjection {
    repeated int source_nums = 1;
    // Empty = identity projection
    MethodDescriptorProto method = 2;
    // Empty = source field name or some construction based on method's output type. Maybe use MethodDescriptorProto or Enum to help resolve custom names?
    string name = 3;
  }
  // These become FieldDescriptor Protos where
  // label, type = MethodDescriptorProto output type
  // name = based on source fields or method
  // number = index
  repeated FieldProjection field_projections = 3;
}
```

However, we actually have another problem: MethodDescriptorProto's input and output types may not have dense field_number representations. This touches on a problem with the proto field/serialization/wire format definition itself: it is fundamentally "sparse" rather than "dense". Of course, this is necessary for proto fields to be able to be deprecated. But since field_nums are ints, this could easily have been "dense-with-holes" rather than "sparse-but-sequential". Eg:

Generally, this a big problem with defining projections and derived protobuf types, because it prevents simple mappings from Lists-of-Fields to function/method parameters. Oftentimes, we'd really rather define a MethodDescriptorProto as MethodName(type1, type2, ... typeN)->(typea, typeb, ...) or directly on protobuf primitives like string, but need to instead wrap them in messages. But, messages as a general class can have arbitrary "holes" where eg field_num 2 is ignored/deprecated. And there is no canonical way to efficiently represent these holes, or the message as a dense range-with-holes rather than sparse map of field_numbers to fields, in a way that is built-in to protobuf itself.

But, we could. It wouldn't be that hard. We'd simply need to version messages and represent them like this

``` proto
message DescriptorProto {
  repeated int deprecated_field_nums = 1;
  repeated DenseFieldDescriptorProto fields = 2; // FieldDescriptorProto where field_number is inferred positionally rather than inline to the Field
  // length(deprecated_field_nums) + length(fields) = max proto field num
  // version num = length(total_field_num).length(deprecated_field_nums)? Each can only increase monotonically
}
```

When a field gets deprecated, we'd append it to the list of deprecated_field_nums and remove it from the list of DenseFieldDescriptorProto. Then, if we ever wanted to construct an instance of this proto from a list of fields, or parse it, we would do so by reading the deprecated_field_nums and "skipping" them. If deprecated_field_nums = [1, 3] and fields has two elements, then we'd interpret the first non-deprecated field as field 2 and the second as field 4. Also, we could represent the wire format as a dense range of sequential fields rather than a "map" of field_nums to values. Although you could construct examples where this is less efficient (eg a huge number of deprecated fields and small number of existing fields), the vast majority of the time, it will be more efficient and compact than the existing map-of-field-nums representation/wire format.

TODO: see how to apply this

# Projecting

Applying a projection is not very complicated: we can take (File)DescriptorProto x Projection -> (File)DescriptorProto | undefined to project an input type to an output type, or google.protobuf.Any x Projection -> google.protobuf.Any | undefined to project an input message to an output message.

``` proto

message ProjectionDescriptorProto {
  string name = 1;
  message FieldProjection {
    repeated int source_nums = 1; // or FieldDescriptorProto
    // Empty = identity projection
    MethodDescriptorProto method = 2;
    // Empty = source field name or some construction based on method's output type. Maybe use MethodDescriptorProto or Enum to help resolve custom names?
    string name = 3; // or FieldDescritproProto
  }
  // These become FieldDescriptor Protos where
  // label, type = MethodDescriptorProto output type
  // name = based on source fields or method
  // number = index
  repeated FieldProjection field_projections = 2;
}

// If deps and def are empty, attempt to resolve type_name (assume it's already registered/resolvable within this context)
message TypeDescriptorProto {
  string name = 1;
  FileDescriptorProto def = 2; // or repeated/inidividual DescriptorProto? but redefined to be inclusive of method/service/enum/field/etc descriptorprotos. idk
  repeated FileDescriptorProto deps = 3;
}

message TypeProjection {
  TypeDescriptorProto source = 1;
  ProjectionDescriptorProto projection = 2;
}

message MessageProjection {
  google.protobuf.Any source = 1; // type_name gets resolved to TypeDescriptorProto
  ProjectionDescriptorProto projection = 2;
}

service Projector {
  rpc ProjectType(TypeProjection) returns(DescriptorProto); // errors if projection is undefined
  rpc ProjectMessage(MessageProjection) returns(google.protobuf.Any); // errors if projection is undefined
}
```

But we may wish to do more complex operations. For example, are there projections from/to a specific type, or between two specific types? Is there a series of projections between two specific types? Can I construct or infer a projection from one type to another based on their fields or some other context? Given a projection, which types can I pass into it, what could be generated from it?

```proto

// representing multiple responses as a stream for now
service Projector {
  // ...
  rpc ProjectionsFrom(TypeDescriptorProto) returns(stream ProjectionDescriptorProto);
  rpc ProjectionsTo(TypeDescriptorProto) returns(stream ProjectionDescriptorProto);
  rpc ProjectionsBetween(stream TypeDescriptorProto) returns(stream ProjectionDescriptorProto);
  rpc DefineProjection(TypeDescriptorProto, TypeDescriptorProto) returns (ProjectionDescriptorProto); // based on field names, types, etc? Maybe ProjectionInferenceOptions?
  rpc ProjectsFrom(ProjectionDescriptorProto) returns(stream TypeDescriptorProto);
  rpc ProjectTo(ProjectionDescriptorProto) returns (stream TypeDescriptorProto);
  rpc ProjectsBetween(ProjectionDescriptorProto) returns (stream TypeDescriptorProto);
}
```

Suppose I have a projection that is defined for any FileDescriptorProto, eg something that Flattens an input projection, or converts all fields to strings according to proto's default string serialization. Should some of these be considered built-int projections? Is conversion from FileDescriptorProto to DescriptorProto to type_name/instance with default values/.proto string definition itself a universal projection? Should we treat invertible projections differently?

Perhaps. Projections are very powerful "functorial" abstractions because they give us a topological/simplical/homotypic representation of protos that allows us to convert between types and their instances without leaving the reflective/abstract world of .proto itself. TBD how we continue to build on this.

# Relational Algebra of Projections

TODO: explain how projections are the building blocks for implementing a relational algebra, similar to how Bigquery and such do so, but in a way that can be generalized across the protobuf ecosystem more generally. Also explain how introducing the Cartesian Product and then representing Joins within that allows us to construct a fully relational algebra from a simpler projection algebra.

Essentially, projections' unique property of "sufficient for describing transformations from T->T' and t->t' for all t in T such that the projection on T is well-defined" maps perfectly to the idea of a Collection being some specific type paired with some number of instances of that type (literally, a type T with multiple t belonging to T).

In other words, a projection allows for conversion between Collection\<T\>->Collection\<T'\> as well as conversion from an identifier (eg a foreign key) to Collection (or type or record within that Collection). That is, projections are the fundamental building blocks we'll use to specify a relational algebra within proto-type.
