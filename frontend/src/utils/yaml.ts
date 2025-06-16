import yaml from "js-yaml";

// Internal helper
const stripEnabled = (obj: any): any => {
	if (Array.isArray(obj)) return obj.map(stripEnabled);

	if (typeof obj === "object" && obj !== null) {
		const result: Record<string, any> = {};
		for (const [key, value] of Object.entries(obj)) {
			if (key === "enabled") continue;
			result[key] = stripEnabled(value);
		}
		return result;
	}

	return obj;
};

// Public API
export const convertToYaml = (input: any): string => {
	try {
		const cleaned = stripEnabled(input);
		const yamlStr = yaml.dump(cleaned, {
			noRefs: true,
			skipInvalid: true,
		});
		return yamlStr;
	} catch (err) {
		console.error("Error converting JSON to YAML:", err);
		return "Invalid JSON";
	}
};
