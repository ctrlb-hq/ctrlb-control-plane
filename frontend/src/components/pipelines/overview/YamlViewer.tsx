import MonacoEditor from "@monaco-editor/react";
import { convertToYaml } from "@/utils/yaml";

const PipelineYAML = ({ jsonforms }: { jsonforms: unknown }) => {
	const yamlOutput = jsonforms == null ? "" : convertToYaml(jsonforms);

	return (
		<div style={{ height: "80vh" }}>
			<MonacoEditor
				height="100%"
				defaultLanguage="yaml"
				value={yamlOutput}
				options={{
					readOnly: true,
					minimap: { enabled: false },
				}}
			/>
		</div>
	);
};

export default PipelineYAML;
