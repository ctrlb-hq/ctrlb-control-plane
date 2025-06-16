import { useEffect, useState } from "react";
import MonacoEditor from "@monaco-editor/react";
import { convertToYaml } from "@/utils/yaml";

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

const PipelineYAML = ({ jsonforms }: { jsonforms: any }) => {
	const [yamlOutput, setYamlOutput] = useState(sampleYaml);

	useEffect(() => {
		const yamlStr = convertToYaml(jsonforms);
		setYamlOutput(yamlStr);
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

export default PipelineYAML;
