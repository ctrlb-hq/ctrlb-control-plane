import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { Button } from "@/components/ui/button";

export function ManifestCodeBlock({
	title,
	code,
	onCopy,
}: {
	title: string;
	code: string;
	onCopy?: () => void;
}) {
	const [copiedCode, setCopiedCode] = useState("");
	const [copyError, setCopyError] = useState(false);
	const copied = copiedCode === code;

	const copy = async () => {
		setCopyError(false);
		try {
			await navigator.clipboard.writeText(code);
			setCopiedCode(code);
			onCopy?.();
		} catch {
			setCopyError(true);
		}
	};

	return (
		<div className="min-w-0 overflow-hidden rounded-lg border">
			<div className="flex items-center justify-between gap-3 bg-gray-50 px-3 py-2">
				<h3 className="text-sm font-medium">{title}</h3>
				<Button
					type="button"
					variant="outline"
					size="sm"
					aria-label={`Copy ${title}`}
					onClick={() => void copy()}
				>
					{copied ? <Check className="mr-2 h-4 w-4" /> : <Copy className="mr-2 h-4 w-4" />}
					{copied ? "Copied" : "Copy"}
				</Button>
			</div>
			<pre
				aria-label={title}
				tabIndex={0}
				className="max-h-80 overflow-auto bg-slate-900 p-3 text-xs leading-relaxed text-slate-100"
			>
				<code>{code}</code>
			</pre>
			{copyError && (
				<p role="alert" className="p-3 text-sm text-red-700">
					Clipboard access failed. Select and copy the text above.
				</p>
			)}
		</div>
	);
}
