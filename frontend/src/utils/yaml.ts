import yaml from "js-yaml";

export const convertToYaml = (input: unknown): string => {
	try {
		return yaml.dump(input, { noRefs: true, skipInvalid: true });
	} catch {
		return "Invalid configuration";
	}
};
