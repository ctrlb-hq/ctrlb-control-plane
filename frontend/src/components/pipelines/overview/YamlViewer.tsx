import { useEffect, useState } from "react";
import yaml from "js-yaml";
import MonacoEditor from "@monaco-editor/react";

const sampleYaml = `
apiVersion: v1
kind: Pod
metadata:
  name: sample-pod
spec:
  containers:
    - name: nginx
      image: nginx:latest
`;

const PipelinYAML = ({ jsonforms }: { jsonforms: any }) => {
	const [yamlOutput, setYamlOutput] = useState(sampleYaml);

	const convertToYaml = () => {
		try {
			const yamlStr = yaml.dump(jsonforms);
			setYamlOutput(yamlStr);
		} catch (error) {
			console.log("Error converting JSON to YAML:", error);
			setYamlOutput("Invalid JSON");
		}
	};
	useEffect(() => {
		convertToYaml();
	}, [jsonforms]);
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

export default PipelinYAML;
