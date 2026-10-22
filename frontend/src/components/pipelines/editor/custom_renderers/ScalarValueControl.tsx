import { useRef } from "react";
import { and, isOneOfControl, rankWith, schemaMatches } from "@jsonforms/core";
import type { CombinatorRendererProps } from "@jsonforms/core";
import { withJsonFormsOneOfProps } from "@jsonforms/react";
import { Unwrapped } from "@jsonforms/material-renderers";

const ScalarValueControl = (props: CombinatorRendererProps) => {
	const variant = useRef(props.indexOfFittingSchema ?? 0);

	// Array rows can move without remounting. Follow the value's type, but keep
	// the current tab when its input is temporarily empty.
	if (props.data !== undefined) {
		variant.current = props.indexOfFittingSchema ?? 0;
	}

	return <Unwrapped.MaterialOneOfRenderer key={variant.current} {...props} />;
};

export const scalarValueRenderer = {
	tester: rankWith(
		4,
		and(
			isOneOfControl,
			schemaMatches(
				schema =>
					schema.oneOf?.length === 3 &&
					["string", "number", "boolean"].every(type =>
						schema.oneOf?.some(option => option.type === type),
					),
			),
		),
	),
	renderer: withJsonFormsOneOfProps(ScalarValueControl),
};

export default ScalarValueControl;
