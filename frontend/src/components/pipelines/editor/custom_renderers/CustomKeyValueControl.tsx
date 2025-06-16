import { withJsonFormsControlProps } from "@jsonforms/react";
import { ControlProps, rankWith, isObjectControl, schemaMatches, and } from "@jsonforms/core";
import { Button, Input } from "@mui/material";
import { Plus } from "lucide-react";

const KeyValueControl = ({ data, handleChange, path, visible = true }: ControlProps) => {
	if (!visible) return null;

	const entries = Object.entries(data || {});

	const updateKey = (oldKey: string, newKey: string) => {
		const updated = { ...data };
		const value = updated[oldKey];
		delete updated[oldKey];
		updated[newKey] = value;
		handleChange(path, updated);
	};

	const updateValue = (key: string, value: string) => {
		const updated = { ...data, [key]: value };
		handleChange(path, updated);
	};

	const removeEntry = (key: string) => {
		const updated = { ...data };
		delete updated[key];
		handleChange(path, updated);
	};

	const addEntry = () => {
		const updated = { ...data };
		const base = "key";
		let i = 1;
		while (updated[`${base}${i}`] !== undefined) i++;
		updated[`${base}${i}`] = "";
		handleChange(path, updated);
	};

	return (
		<div className="flex flex-col gap-2">
			{entries.map(([key, value], index) => (
				<div key={index} className="flex gap-2 items-center">
					<Input value={key} onChange={e => updateKey(key, e.target.value)} placeholder="Header Key" />
					<Input
						value={value}
						onChange={e => updateValue(key, e.target.value)}
						placeholder="Header Value"
					/>
					<Button type="button" onClick={() => removeEntry(key)}></Button>
				</div>
			))}
			<Button type="button" onClick={addEntry}>
				<Plus/> Add Header
			</Button>
		</div>
	);
};

const headersTester = rankWith(
	5,
	and(
		isObjectControl,
		schemaMatches(
			schema =>
				schema?.title === "Custom Headers" &&
				schema?.type === "object" &&
				!!schema?.additionalProperties,
		),
	),
);

export const customKeyValueRenderer = {
	tester: headersTester,
	renderer: withJsonFormsControlProps(KeyValueControl),
};

export default KeyValueControl;
