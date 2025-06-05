package v1alpha1

import (
	"time"
	"net"
)

@jsonschema(schema="http://json-schema.org/draft-07/schema#")
@jsonschema(id="http://csrc.nist.gov/ns/oscal/1.1.3/oscal-component-definition-schema.json")
close({
	$schema?:                #."json-schema-directive"
	"component-definition"!: #."oscal-component-definition-oscal-component-definition:component-definition"
})

#: "json-schema-directive": #URIReferenceDatatype

// Capability
//
// A grouping of other components and/or capabilities.
#: "oscal-component-definition-oscal-component-definition:capability": close({
	uuid!: #UUIDDatatype
	name!: #StringDatatype

	// Capability Description
	//
	// A summary of the capability.
	description!: string
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	"incorporates-components"?: [_, ...] & [...#."oscal-component-definition-oscal-component-definition:incorporates-component"]
	"control-implementations"?: [_, ...] & [...#."oscal-component-definition-oscal-component-definition:control-implementation"]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// Component Definition
//
// A collection of component descriptions, which may optionally be
// grouped by capability.
#: "oscal-component-definition-oscal-component-definition:component-definition": close({
	uuid!:     #UUIDDatatype
	metadata!: #."oscal-component-definition-oscal-metadata:metadata"
	"import-component-definitions"?: [_, ...] & [...#."oscal-component-definition-oscal-component-definition:import-component-definition"]
	components?: [_, ...] & [...#."oscal-component-definition-oscal-component-definition:defined-component"]
	capabilities?: [_, ...] & [...#."oscal-component-definition-oscal-component-definition:capability"]
	"back-matter"?: #."oscal-component-definition-oscal-metadata:back-matter"
})

// Control Implementation Set
//
// Defines how the component or capability supports a set of
// controls.
#: "oscal-component-definition-oscal-component-definition:control-implementation": close({
	uuid!:   #UUIDDatatype
	source!: #URIReferenceDatatype

	// Control Implementation Description
	//
	// A description of how the specified set of controls are
	// implemented for the containing component or capability.
	description!: string
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	"set-parameters"?: [_, ...] & [...#."oscal-component-definition-oscal-implementation-common:set-parameter"]
	"implemented-requirements"!: [_, ...] & [...#."oscal-component-definition-oscal-component-definition:implemented-requirement"]
})

// Component
//
// A defined component that can be part of an implemented system.
#: "oscal-component-definition-oscal-component-definition:defined-component": close({
	uuid!: #UUIDDatatype

	// Component Type
	//
	// A category describing the purpose of the component.
	type!: matchN(>=1, [#StringDatatype, "interconnection" | "software" | "hardware" | "service" | "policy" | "physical" | "process-procedure" | "plan" | "guidance" | "standard" | "validation"])

	// Component Title
	//
	// A human readable name for the component.
	title!: string

	// Component Description
	//
	// A description of the component, including information about its
	// function.
	description!: string

	// Purpose
	//
	// A summary of the technological or business purpose of the
	// component.
	purpose?: string
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	"responsible-roles"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:responsible-role"]
	protocols?: [_, ...] & [...#."oscal-component-definition-oscal-implementation-common:protocol"]
	"control-implementations"?: [_, ...] & [...#."oscal-component-definition-oscal-component-definition:control-implementation"]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// Control Implementation
//
// Describes how the containing component or capability implements
// an individual control.
#: "oscal-component-definition-oscal-component-definition:implemented-requirement": close({
	uuid!:         #UUIDDatatype
	"control-id"!: #TokenDatatype

	// Control Implementation Description
	//
	// A suggestion from the supplier (e.g., component vendor or
	// author) for how the specified control may be implemented if
	// the containing component or capability is instantiated in a
	// system security plan.
	description!: string
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	"set-parameters"?: [_, ...] & [...#."oscal-component-definition-oscal-implementation-common:set-parameter"]
	"responsible-roles"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:responsible-role"]
	statements?: [_, ...] & [...#."oscal-component-definition-oscal-component-definition:statement"]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// Import Component Definition
//
// Loads a component definition from another resource.
#: "oscal-component-definition-oscal-component-definition:import-component-definition": close({
	href!: #URIReferenceDatatype
})

// Incorporates Component
//
// The collection of components comprising this capability.
#: "oscal-component-definition-oscal-component-definition:incorporates-component": close({
	"component-uuid"!: #UUIDDatatype

	// Component Description
	//
	// A description of the component, including information about its
	// function.
	description!: string
})

// Control Statement Implementation
//
// Identifies which statements within a control are addressed.
#: "oscal-component-definition-oscal-component-definition:statement": close({
	"statement-id"!: #TokenDatatype
	uuid!:           #UUIDDatatype

	// Statement Implementation Description
	//
	// A summary of how the containing control statement is
	// implemented by the component or capability.
	description!: string
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	"responsible-roles"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:responsible-role"]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// Include All
//
// Include all controls from the imported catalog or profile
// resources.
#: "oscal-component-definition-oscal-control-common:include-all": close({})

// Parameter
//
// Parameters provide a mechanism for the dynamic assignment of
// value(s) in a control.
#: "oscal-component-definition-oscal-control-common:parameter": close({
	id!:           #TokenDatatype
	class?:        #TokenDatatype
	"depends-on"?: #TokenDatatype
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]

	// Parameter Label
	//
	// A short, placeholder name for the parameter, which can be used
	// as a substitute for a value if no value is assigned.
	label?: string

	// Parameter Usage Description
	//
	// Describes the purpose and use of a parameter.
	usage?: string
	constraints?: [_, ...] & [...#."oscal-component-definition-oscal-control-common:parameter-constraint"]
	guidelines?: [_, ...] & [...#."oscal-component-definition-oscal-control-common:parameter-guideline"]
	values?: [_, ...] & [...#."oscal-component-definition-oscal-control-common:parameter-value"]
	select?:  #."oscal-component-definition-oscal-control-common:parameter-selection"
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// Constraint
//
// A formal or informal expression of a constraint or test.
#: "oscal-component-definition-oscal-control-common:parameter-constraint": close({
	// Constraint Description
	//
	// A textual summary of the constraint to be applied.
	description?: string
	tests?: [_, ...] & [...close({
		expression!: #StringDatatype
		remarks?:    #."oscal-component-definition-oscal-metadata:remarks"
	})]
})

// Guideline
//
// A prose statement that provides a recommendation for the use of
// a parameter.
#: "oscal-component-definition-oscal-control-common:parameter-guideline": close({
	// Guideline Text
	//
	// Prose permits multiple paragraphs, lists, tables etc.
	prose!: string
})

// Selection
//
// Presenting a choice among alternatives.
#: "oscal-component-definition-oscal-control-common:parameter-selection": close({
	// Parameter Cardinality
	//
	// Describes the number of selections that must occur. Without
	// this setting, only one value should be assumed to be
	// permitted.
	"how-many"?: matchN(2, [#TokenDatatype, "one" | "one-or-more"]) & string
	choice?: [_, ...] & [...string]
})

#: "oscal-component-definition-oscal-control-common:parameter-value": #StringDatatype

// Part
//
// An annotated, markup-based textual element of a control's or
// catalog group's definition, or a child of another part.
#: "oscal-component-definition-oscal-control-common:part": close({
	id?:    #TokenDatatype
	name!:  #TokenDatatype
	ns?:    #URIDatatype
	class?: #TokenDatatype

	// Part Title
	//
	// An optional name given to the part, which may be used by a tool
	// for display and navigation.
	title?: string
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]

	// Part Text
	//
	// Permits multiple paragraphs, lists, tables etc.
	prose?: string
	parts?: [_, ...] & [...#."oscal-component-definition-oscal-control-common:part"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
})

// Privilege
//
// Identifies a specific system privilege held by the user, along
// with an associated description and/or rationale for the
// privilege.
#: "oscal-component-definition-oscal-implementation-common:authorized-privilege": close({
	// Privilege Title
	//
	// A human readable name for the privilege.
	title!: string

	// Privilege Description
	//
	// A summary of the privilege's purpose within the system.
	description?: string
	"functions-performed"!: [_, ...] & [...#."oscal-component-definition-oscal-implementation-common:function-performed"]
})

#: "oscal-component-definition-oscal-implementation-common:function-performed": #StringDatatype

// Implementation Status
//
// Indicates the degree to which the a given control is
// implemented.
#: "oscal-component-definition-oscal-implementation-common:implementation-status": close({
	// Implementation State
	//
	// Identifies the implementation status of the control or control
	// objective.
	state!: matchN(>=1, [#TokenDatatype, "implemented" | "partial" | "planned" | "alternative" | "not-applicable"])
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// Inventory Item
//
// A single managed inventory item within the system.
#: "oscal-component-definition-oscal-implementation-common:inventory-item": close({
	uuid!: #UUIDDatatype

	// Inventory Item Description
	//
	// A summary of the inventory item stating its purpose within the
	// system.
	description!: string
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	"responsible-parties"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:responsible-party"]
	"implemented-components"?: [_, ...] & [...close({
		"component-uuid"!: #UUIDDatatype
		props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
		links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
		"responsible-parties"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:responsible-party"]
		remarks?: #."oscal-component-definition-oscal-metadata:remarks"
	})]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// Port Range
//
// Where applicable this is the transport layer protocol port
// range an IPv4-based or IPv6-based service uses.
#: "oscal-component-definition-oscal-implementation-common:port-range": close({
	start?: #NonNegativeIntegerDatatype
	end?:   #NonNegativeIntegerDatatype

	// Transport
	//
	// Indicates the transport type.
	transport?: matchN(2, [#TokenDatatype, "TCP" | "UDP"]) & string
})

// Service Protocol Information
//
// Information about the protocol used to provide a service.
#: "oscal-component-definition-oscal-implementation-common:protocol": close({
	uuid?: #UUIDDatatype
	name?: #StringDatatype

	// Protocol Title
	//
	// A human readable name for the protocol (e.g., Transport Layer
	// Security).
	title?: string
	"port-ranges"?: [_, ...] & [...#."oscal-component-definition-oscal-implementation-common:port-range"]
})

// Set Parameter Value
//
// Identifies the parameter that will be set by the enclosed
// value.
#: "oscal-component-definition-oscal-implementation-common:set-parameter": close({
	"param-id"!: #TokenDatatype
	values!: [_, ...] & [...#StringDatatype]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// Component
//
// A defined component that can be part of an implemented system.
#: "oscal-component-definition-oscal-implementation-common:system-component": close({
	uuid!: #UUIDDatatype

	// Component Type
	//
	// A category describing the purpose of the component.
	type!: matchN(>=1, [#StringDatatype, "this-system" | "system" | "interconnection" | "software" | "hardware" | "service" | "policy" | "physical" | "process-procedure" | "plan" | "guidance" | "standard" | "validation" | "network"])

	// Component Title
	//
	// A human readable name for the system component.
	title!: string

	// Component Description
	//
	// A description of the component, including information about its
	// function.
	description!: string

	// Purpose
	//
	// A summary of the technological or business purpose of the
	// component.
	purpose?: string
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]

	// Status
	//
	// Describes the operational status of the system component.
	status!: close({
		// State
		//
		// The operational status.
		state!: matchN(2, [#TokenDatatype, "under-development" | "operational" | "disposition" | "other"]) & string
		remarks?: #."oscal-component-definition-oscal-metadata:remarks"
	})
	"responsible-roles"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:responsible-role"]
	protocols?: [_, ...] & [...#."oscal-component-definition-oscal-implementation-common:protocol"]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// System Identification
//
// A human-oriented, globally unique identifier with
// cross-instance scope that can be used to reference this system
// identification property elsewhere in this or other OSCAL
// instances. When referencing an externally defined system
// identification, the system identification must be used in the
// context of the external / imported OSCAL instance (e.g.,
// uri-reference). This string should be assigned per-subject,
// which means it should be consistently used to identify the
// same system across revisions of the document.
#: "oscal-component-definition-oscal-implementation-common:system-id": close({
	// Identification System Type
	//
	// Identifies the identification system from which the provided
	// identifier was assigned.
	"identifier-type"?: matchN(>=1, [#URIDatatype, "https://fedramp.gov" | "http://fedramp.gov/ns/oscal" | "https://ietf.org/rfc/rfc4122" | "http://ietf.org/rfc/rfc4122"])
	id!: #StringDatatype
})

// System User
//
// A type of user that interacts with the system based on an
// associated role.
#: "oscal-component-definition-oscal-implementation-common:system-user": close({
	uuid!: #UUIDDatatype

	// User Title
	//
	// A name given to the user, which may be used by a tool for
	// display and navigation.
	title?:        string
	"short-name"?: #StringDatatype

	// User Description
	//
	// A summary of the user's purpose within the system.
	description?: string
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	"role-ids"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:role-id"]
	"authorized-privileges"?: [_, ...] & [...#."oscal-component-definition-oscal-implementation-common:authorized-privilege"]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// Action
//
// An action applied by a role within a given party to the
// content.
#: "oscal-component-definition-oscal-metadata:action": close({
	uuid!:   #UUIDDatatype
	date?:   #DateTimeWithTimezoneDatatype
	type!:   #TokenDatatype
	system!: #URIDatatype
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	"responsible-parties"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:responsible-party"]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

#: "oscal-component-definition-oscal-metadata:addr-line": #StringDatatype

// Address
//
// A postal address for the location.
#: "oscal-component-definition-oscal-metadata:address": close({
	// Address Type
	//
	// Indicates the type of address.
	type?: matchN(>=1, [#TokenDatatype, "home" | "work"])
	"addr-lines"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:addr-line"]
	city?:          #StringDatatype
	state?:         #StringDatatype
	"postal-code"?: #StringDatatype
	country?:       #StringDatatype
})

// Back matter
//
// A collection of resources that may be referenced from within
// the OSCAL document instance.
#: "oscal-component-definition-oscal-metadata:back-matter": close({
	resources?: [_, ...] & [...close({
		uuid!: #UUIDDatatype

		// Resource Title
		//
		// An optional name given to the resource, which may be used by a
		// tool for display and navigation.
		title?: string

		// Resource Description
		//
		// An optional short summary of the resource used to indicate the
		// purpose of the resource.
		description?: string
		props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
		"document-ids"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:document-id"]

		// Citation
		//
		// An optional citation consisting of end note text using
		// structured markup.
		citation?: close({
			// Citation Text
			//
			// A line of citation text.
			text!: string
			props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
			links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
		})
		rlinks?: [_, ...] & [...close({
			href!:         #URIReferenceDatatype
			"media-type"?: #StringDatatype
			hashes?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:hash"]
		})]

		// Base64
		//
		// A resource encoded using the Base64 alphabet defined by RFC
		// 2045.
		base64?: close({
			filename?:     #TokenDatatype
			"media-type"?: #StringDatatype
			value!:        #Base64Datatype
		})
		remarks?: #."oscal-component-definition-oscal-metadata:remarks"
	})]
})

// Document Identifier
//
// A document identifier qualified by an identifier scheme.
#: "oscal-component-definition-oscal-metadata:document-id": close({
	// Document Identification Scheme
	//
	// Qualifies the kind of document identifier using a URI. If the
	// scheme is not provided the value of the element will be
	// interpreted as a string of characters.
	scheme?: matchN(>=1, [#URIDatatype, "http://www.doi.org/"])
	identifier!: #StringDatatype
})

#: "oscal-component-definition-oscal-metadata:email-address": #EmailAddressDatatype

// Hash
//
// A representation of a cryptographic digest generated over a
// resource using a specified hash algorithm.
#: "oscal-component-definition-oscal-metadata:hash": close({
	// Hash algorithm
	//
	// The digest method by which a hash is derived.
	algorithm!: matchN(>=1, [#StringDatatype, "SHA-224" | "SHA-256" | "SHA-384" | "SHA-512" | "SHA3-224" | "SHA3-256" | "SHA3-384" | "SHA3-512"])
	value!: #StringDatatype
})

#: "oscal-component-definition-oscal-metadata:last-modified": #DateTimeWithTimezoneDatatype

// Link
//
// A reference to a local or remote resource, that has a specific
// relation to the containing object.
#: "oscal-component-definition-oscal-metadata:link": close({
	href!: #URIReferenceDatatype

	// Link Relation Type
	//
	// Describes the type of relationship provided by the link's
	// hypertext reference. This can be an indicator of the link's
	// purpose.
	rel?: matchN(>=1, [#TokenDatatype, "reference"])
	"media-type"?:        #StringDatatype
	"resource-fragment"?: #StringDatatype

	// Link Text
	//
	// A textual label to associate with the link, which may be used
	// for presentation in a tool.
	text?: string
})

#: "oscal-component-definition-oscal-metadata:location-uuid": #UUIDDatatype

// Document Metadata
//
// Provides information about the containing document, and defines
// concepts that are shared across the document.
#: "oscal-component-definition-oscal-metadata:metadata": close({
	// Document Title
	//
	// A name given to the document, which may be used by a tool for
	// display and navigation.
	title!:           string
	published?:       #."oscal-component-definition-oscal-metadata:published"
	"last-modified"!: #."oscal-component-definition-oscal-metadata:last-modified"
	version!:         #."oscal-component-definition-oscal-metadata:version"
	"oscal-version"!: #."oscal-component-definition-oscal-metadata:oscal-version"
	revisions?: [_, ...] & [...close({
		// Document Title
		//
		// A name given to the document revision, which may be used by a
		// tool for display and navigation.
		title?:           string
		published?:       #."oscal-component-definition-oscal-metadata:published"
		"last-modified"?: #."oscal-component-definition-oscal-metadata:last-modified"
		version!:         #."oscal-component-definition-oscal-metadata:version"
		"oscal-version"?: #."oscal-component-definition-oscal-metadata:oscal-version"
		props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
		links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
		remarks?: #."oscal-component-definition-oscal-metadata:remarks"
	})]
	"document-ids"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:document-id"]
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	roles?: [_, ...] & [...close({
		id!: #TokenDatatype

		// Role Title
		//
		// A name given to the role, which may be used by a tool for
		// display and navigation.
		title!:        string
		"short-name"?: #StringDatatype

		// Role Description
		//
		// A summary of the role's purpose and associated
		// responsibilities.
		description?: string
		props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
		links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
		remarks?: #."oscal-component-definition-oscal-metadata:remarks"
	})]
	locations?: [_, ...] & [...close({
		uuid!: #UUIDDatatype

		// Location Title
		//
		// A name given to the location, which may be used by a tool for
		// display and navigation.
		title?:   string
		address?: #."oscal-component-definition-oscal-metadata:address"
		"email-addresses"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:email-address"]
		"telephone-numbers"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:telephone-number"]
		urls?: [_, ...] & [...#URIDatatype]
		props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
		links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
		remarks?: #."oscal-component-definition-oscal-metadata:remarks"
	})]
	parties?: [_, ...] & [...close({
		uuid!: #UUIDDatatype

		// Party Type
		//
		// A category describing the kind of party the object describes.
		type!: matchN(2, [#StringDatatype, "person" | "organization"]) & string
		name?:         #StringDatatype
		"short-name"?: #StringDatatype
		"external-ids"?: [_, ...] & [...close({
			// External Identifier Schema
			//
			// Indicates the type of external identifier.
			scheme!: matchN(>=1, [#URIDatatype, "http://orcid.org/"])
			id!: #StringDatatype
		})]
		props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
		links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
		"email-addresses"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:email-address"]
		"telephone-numbers"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:telephone-number"]
		addresses?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:address"]
		"location-uuids"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:location-uuid"]
		"member-of-organizations"?: [_, ...] & [...#UUIDDatatype]
		remarks?: #."oscal-component-definition-oscal-metadata:remarks"
	})]
	"responsible-parties"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:responsible-party"]
	actions?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:action"]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

#: "oscal-component-definition-oscal-metadata:oscal-version": #StringDatatype

#: "oscal-component-definition-oscal-metadata:party-uuid": #UUIDDatatype

// Property
//
// An attribute, characteristic, or quality of the containing
// object expressed as a namespace qualified name/value pair.
#: "oscal-component-definition-oscal-metadata:property": close({
	name!:    #TokenDatatype
	uuid?:    #UUIDDatatype
	ns?:      #URIDatatype
	value!:   #StringDatatype
	class?:   #TokenDatatype
	group?:   #TokenDatatype
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

#: "oscal-component-definition-oscal-metadata:published": #DateTimeWithTimezoneDatatype

// Remarks
//
// Additional commentary about the containing object.
#: "oscal-component-definition-oscal-metadata:remarks": string

// Responsible Party
//
// A reference to a set of persons and/or organizations that have
// responsibility for performing the referenced role in the
// context of the containing object.
#: "oscal-component-definition-oscal-metadata:responsible-party": close({
	"role-id"!: #TokenDatatype
	"party-uuids"!: [_, ...] & [...#."oscal-component-definition-oscal-metadata:party-uuid"]
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

// Responsible Role
//
// A reference to a role with responsibility for performing a
// function relative to the containing object, optionally
// associated with a set of persons and/or organizations that
// perform that role.
#: "oscal-component-definition-oscal-metadata:responsible-role": close({
	"role-id"!: #TokenDatatype
	props?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:property"]
	links?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:link"]
	"party-uuids"?: [_, ...] & [...#."oscal-component-definition-oscal-metadata:party-uuid"]
	remarks?: #."oscal-component-definition-oscal-metadata:remarks"
})

#: "oscal-component-definition-oscal-metadata:role-id": #TokenDatatype

// Telephone Number
//
// A telephone service number as defined by ITU-T E.164.
#: "oscal-component-definition-oscal-metadata:telephone-number": close({
	// type flag
	//
	// Indicates the type of phone number.
	type?: matchN(>=1, [#StringDatatype, "home" | "office" | "mobile"])
	number!: #StringDatatype
})

#: "oscal-component-definition-oscal-metadata:version": #StringDatatype

// Binary data encoded using the Base 64 encoding algorithm as
// defined by RFC4648.
#Base64Datatype: =~"^[0-9A-Za-z+/]+={0,2}$"

// A string representing a point in time with a required timezone.
#DateTimeWithTimezoneDatatype: time.Time & =~"^(((2000|2400|2800|(19|2[0-9](0[48]|[2468][048]|[13579][26])))-02-29)|(((19|2[0-9])[0-9]{2})-02-(0[1-9]|1[0-9]|2[0-8]))|(((19|2[0-9])[0-9]{2})-(0[13578]|10|12)-(0[1-9]|[12][0-9]|3[01]))|(((19|2[0-9])[0-9]{2})-(0[469]|11)-(0[1-9]|[12][0-9]|30)))T(2[0-3]|[01][0-9]):([0-5][0-9]):([0-5][0-9])(\\.[0-9]+)?(Z|(-((0[0-9]|1[0-2]):00|0[39]:30)|\\+((0[0-9]|1[0-4]):00|(0[34569]|10):30|(0[58]|12):45)))$"

// An email address string formatted according to RFC 6531.
#EmailAddressDatatype: matchN(2, [#StringDatatype, =~"^.+@.+$"]) & string

// A whole number value.
#IntegerDatatype: int

// An integer value that is equal to or greater than 0.
#NonNegativeIntegerDatatype: matchN(2, [#IntegerDatatype, >=0]) & number

// A non-empty string with leading and trailing whitespace
// disallowed. Whitespace is: U+9, U+10, U+32 or [
// ]+
#StringDatatype: =~"^\\S(.*\\S)?$"

// A non-colonized name as defined by XML Schema Part 2: Datatypes
// Second Edition. https://www.w3.org/TR/xmlschema11-2/#NCName.
#TokenDatatype: =~"^(\\p{L}|_)(\\p{L}|\\p{N}|[.\\-_])*$"

// A universal resource identifier (URI) formatted according to
// RFC3986.
#URIDatatype: net.AbsURL & =~"^[a-zA-Z][a-zA-Z0-9+\\-.]+:.+$"

// A URI Reference, either a URI or a relative-reference,
// formatted according to section 4.1 of RFC3986.
#URIReferenceDatatype: net.URL

// A type 4 ('random' or 'pseudorandom') or type 5 UUID per RFC
// 4122.
#UUIDDatatype: =~"^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[45][0-9A-Fa-f]{3}-[89ABab][0-9A-Fa-f]{3}-[0-9A-Fa-f]{12}$"
