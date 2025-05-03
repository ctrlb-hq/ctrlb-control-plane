import { Editor } from "@monaco-editor/react";

const EditPipelineYAML = () => {
	const samplePipelineYAML = `
version: '1.0'

stages:
  - name: Build
    jobs:
      - name: Compile
        steps:
          - name: Checkout code
            uses: actions/checkout@v2
          - name: Set up Node.js
            uses: actions/setup-node@v2
            with:
              node-version: '14'
          - name: Install dependencies
            run: npm install
          - name: Run build         
            run: npm run build

  - name: Test
    jobs:
      - name: Unit Tests
        steps:
          - name: Checkout code
            uses: actions/checkout@v2
          - name: Set up Node.js
            uses: actions/setup-node@v2
            with:
              node-version: '14'
          - name: Install dependencies
            run: npm install
          - name: Run tests
            run: npm test

  - name: Deploy
    jobs:
      - name: Deploy to Production
        steps:
          - name: Checkout code
            uses: actions/checkout@v2
          - name: Set up Node.js
            uses: actions/setup-node@v2
            with:
              node-version: '14'
          - name: Install dependencies
            run: npm install
          - name: Deploy
            run: npm run deploy
  `;

	return (
		<div>
			<Editor height="80vh" defaultLanguage="yaml" defaultValue={samplePipelineYAML} />
		</div>
	);
};

export default EditPipelineYAML;
