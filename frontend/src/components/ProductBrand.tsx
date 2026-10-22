interface ProductBrandProps {
	variant?: "auth" | "header";
}

export default function ProductBrand({ variant = "header" }: ProductBrandProps) {
	const centered = variant === "auth";

	return (
		<header
			className={centered ? "flex flex-col items-center gap-3 text-center" : "flex items-start gap-3"}
		>
			<img
				src={`${import.meta.env.BASE_URL}ctrlb-logo.png`}
				alt=""
				width={48}
				height={48}
				className={centered ? "h-12 w-12 shrink-0 object-contain" : "h-10 w-10 shrink-0 object-contain"}
			/>
			<div className="min-w-0 space-y-1">
				<h1
					className={`font-semibold tracking-tight text-gray-900 ${
						centered ? "text-2xl sm:text-3xl" : "text-xl sm:text-2xl"
					}`}
				>
					CtrlB Control Plane
				</h1>
				<p className="text-sm leading-6 text-gray-600">
					Manage your OpenTelemetry collectors and pipelines in one place.
				</p>
			</div>
		</header>
	);
}
