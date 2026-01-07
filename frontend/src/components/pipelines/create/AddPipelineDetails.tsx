import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { AlertCircle, CopyIcon, Loader2, BadgeCheck } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import ProgressFlow from "@/components/pipelines/create/ProgressFlow";
import agentServices from "@/services/agent";
import { installCommands, installCommandsFluentBit } from "@/constants";
import { useGlobalSnackbar } from "@/context/useGlobalSnackbar";

type Platform = "linux" | "macOS" | "kubernetes" | "openShift";
type AgentType = "otel" | "fluent-bit";

interface formData {
	name: string;
	platform: Platform | "";
	agentType: AgentType;
}

interface AddPipelineDetailsProps {
	sendPipelineDataToParent: (id: string, name: string) => void;
	currentStep: number;
	setCurrentStep: React.Dispatch<React.SetStateAction<number>>;
}

const AddPipelineDetails = ({
	sendPipelineDataToParent,
	currentStep,
	setCurrentStep,
}: AddPipelineDetailsProps) => {
	const pipelineName = "";
	const platform = null;

	const [showRunCommand, setShowRunCommand] = useState(false);
	const [showHeartBeat, setShowHeartBeat] = useState(false);
	const [showStatus, setShowStatus] = useState(false);
	const [status, setStatus] = useState<"success" | "failed">("failed");
	const [_showAgentInfo, setShowAgentInfo] = useState(false);
	const [_isApiKeyCopied, setIsApiKeyCopied] = useState(false);
	const [showConfigureButton, setShowConfigureButton] = useState(false);
	const [_isChecking, setIsChecking] = useState(false);
	const abortControllerRef = useRef<AbortController | null>(null);
	const { showSnackbar } = useGlobalSnackbar();
	const [formData, setFormData] = useState<formData>({
		name: pipelineName ?? "",
		platform: platform ?? "",
		agentType: "otel",
	});

	const [errors, setErrors] = useState({
		name: false,
		platform: false,
		agentType: false,
	});

	const [touched, setTouched] = useState({
		name: false,
		platform: false,
		agentType: false,
	});

	const handleChange = (e: any) => {
		const { id, value } = e.target;
		setFormData(prev => ({
			...prev,
			[id]: value,
		}));
		// Clear error when user types
		if (value.trim()) {
			setErrors(prev => ({
				...prev,
				[id]: false,
			}));
		}
	};

	const handleSubmit = (e: any) => {
		e.preventDefault();
		// Check required fields
		const newErrors = {
			name: !formData.name.trim(),
			platform: !formData.platform,
			agentType: !formData.agentType,
		};

		setErrors(newErrors);
		setTouched({
			name: true,
			platform: true,
			agentType: true,
		});

		setShowRunCommand(true);
	};

	const handleCopy = async (command: string) => {
		try {
			await navigator.clipboard.writeText(command);
			showSnackbar("Install command copied to clipboard", "success", 2000);
			setIsApiKeyCopied(true);
		} catch (error) {
			console.error("Clipboard copy failed:", error);
			showSnackbar(
				"Clipboard access blocked. Press ⌘ + C (Mac) or Ctrl + C (Windows) to copy manually.",
				"error",
				3000
			);
		}

		const since = Math.floor(new Date().getTime() / 1000);
		setShowConfigureButton(true);

		setTimeout(() => setShowHeartBeat(true), 2000);
		setTimeout(() => setShowStatus(true), 6000);
		setTimeout(() => {
			setShowAgentInfo(true);
			checkAgentStatus(since);
		}, 1000);
	};

	const handleTryAgain = () => {
		setShowStatus(false);
		setStatus("failed");
		setShowHeartBeat(false);

		setTimeout(() => {
			const since = Math.floor(new Date().getTime() / 1000);
			checkAgentStatus(since);
		}, 1000);
	};

	const stopChecking = useCallback(() => {
		if (abortControllerRef.current) {
			abortControllerRef.current.abort();
			abortControllerRef.current = null;
		}
		setIsChecking(false);
	}, []);

	useEffect(() => {
		return () => {
			stopChecking();
		};
	}, [stopChecking]);

	const checkAgentStatus = async (since: number) => {
		// Stop any existing check
		stopChecking();
		// Create new abort controller
		const abortController = new AbortController();
		abortControllerRef.current = abortController;
		setIsChecking(true);
		setShowHeartBeat(true);
		setShowStatus(false);

		const THREE_MINUTES = 3 * 60 * 1000;
		const CHECK_INTERVAL = 3 * 1000;
		const startTime = Date.now();

		try {
			while (!abortController.signal.aborted) {
				try {
					// Check if we've exceeded the time limit
					if (Date.now() - startTime >= THREE_MINUTES) {
						setStatus("failed");
						setShowStatus(true);
						setShowHeartBeat(false);
						stopChecking();
						break;
					}

					const data = await agentServices.getLatestAgents({ since });
					if (data) {
						setStatus(data ? "success" : "failed");
						setShowStatus(true);
						setShowHeartBeat(false);
						stopChecking();
						sendPipelineDataToParent(data?.pipeline_id, formData.name);
					}

					await new Promise(resolve => setTimeout(resolve, CHECK_INTERVAL));
				} catch (error) {
					if (abortController.signal.aborted) {
						break;
					}
					console.error("Error checking agents:", error);
				}
			}
		} catch (error) {
			console.error("Error in checkAgentStatus:", error);
			if (!abortController.signal.aborted) {
				setStatus("failed");
				setShowStatus(true);
				setShowHeartBeat(false);
			}
		} finally {
			if (abortController === abortControllerRef.current) {
				stopChecking();
			}
		}
	};

	return (
		<div className="flex flex-row gap-5 mt-4">
			<div className="w-1/4 h-full">
				<ProgressFlow currentStep={currentStep} />
			</div>
			<Card className="w-3/4 h-full">
				<CardHeader>
					<CardTitle className="text-xl font-bold">Let's get started building your Pipeline.</CardTitle>

					<p className="text-gray-600 mt-2">Let's get started building your pipeline configuration.</p>
				</CardHeader>
				<CardContent className="h-auto min-h-[37rem]">
					<form className="space-y-6" onSubmit={handleSubmit}>
						<div className="space-y-2">
							<Label htmlFor="name" className="text-base font-medium flex items-center">
								Name <span className="text-red-500 ml-1">*</span>
							</Label>
							<Input
								id="name"
								value={formData.name}
								onChange={handleChange}
								// onBlur is not supported by Select
								className={`h-10 ${errors.name && touched.name ? "border-red-500 focus-visible:ring-red-500" : "border-gray-300"}`}
								required
							/>
							{errors.name && touched.name && (
								<div className="flex items-center mt-1 text-red-500 text-sm">
									<AlertCircle className="w-4 h-4 mr-1" />
									<span>Name is required</span>
								</div>
							)}
						</div>
						<div className="space-y-2">
							<Label htmlFor="platform" className="text-base font-medium flex items-center">
								Platform <span className="text-red-500 ml-1">*</span>
							</Label>
							<select
								id="platform"
								value={formData.platform}
								onChange={e => {
									const value = e.target.value as Platform;
									setFormData(prev => ({ ...prev, platform: value }));

									// Clear error when user selects
									if (value.length > 0) {
										setErrors(prev => ({
											...prev,
											platform: false,
										}));
									}
								}}
								className={`h-10 w-full border rounded-md px-3 py-2 bg-white text-sm ${errors.platform && touched.platform ? "border-red-500 focus:ring-red-500" : "border-gray-300 focus:ring-blue-500"}`}>
								<option value="" disabled>
									Select a platform
								</option>
								<option value="linux">Linux</option>
								<option value="kubernetes">Kubernetes</option>
								<option value="macOS">macOS</option>
							</select>

							{errors.platform && touched.platform && (
								<div className="flex items-center mt-1 text-red-500 text-sm">
									<AlertCircle className="w-4 h-4 mr-1" />
									<span>At least one platform must be selected</span>
								</div>
							)}

							{errors.name && touched.name && (
								<div className="flex items-center mt-1 text-red-500 text-sm">
									<AlertCircle className="w-4 h-4 mr-1" />
									<span>Platform is required</span>
								</div>
							)}
						</div>
						<div className="space-y-2">
							<Label htmlFor="agentType" className="text-base font-medium flex items-center">
								Agent Type <span className="text-red-500 ml-1">*</span>
							</Label>
							<div className="flex gap-4">
								<label
									className={`flex-1 flex items-center justify-center gap-2 p-4 border-2 rounded-lg cursor-pointer transition-all ${
										formData.agentType === "otel"
											? "border-blue-500 bg-blue-50"
											: "border-gray-200 hover:border-gray-300"
									}`}>
									<input
										type="radio"
										name="agentType"
										value="otel"
										checked={formData.agentType === "otel"}
										onChange={() => {
											setFormData(prev => ({ ...prev, agentType: "otel" }));
											setErrors(prev => ({ ...prev, agentType: false }));
										}}
										className="sr-only"
									/>
									<div className="text-center">
										<div className="font-semibold text-gray-800">OpenTelemetry</div>
										<div className="text-xs text-gray-500">OTEL Collector</div>
									</div>
								</label>
								<label
									className={`flex-1 flex items-center justify-center gap-2 p-4 border-2 rounded-lg cursor-pointer transition-all ${
										formData.agentType === "fluent-bit"
											? "border-blue-500 bg-blue-50"
											: "border-gray-200 hover:border-gray-300"
									}`}>
									<input
										type="radio"
										name="agentType"
										value="fluent-bit"
										checked={formData.agentType === "fluent-bit"}
										onChange={() => {
											setFormData(prev => ({ ...prev, agentType: "fluent-bit" }));
											setErrors(prev => ({ ...prev, agentType: false }));
										}}
										className="sr-only"
									/>
									<div className="text-center">
										<div className="font-semibold text-gray-800">Fluent Bit</div>
										<div className="text-xs text-gray-500">Log processor</div>
									</div>
								</label>
							</div>
							{errors.agentType && touched.agentType && (
								<div className="flex items-center mt-1 text-red-500 text-sm">
									<AlertCircle className="w-4 h-4 mr-1" />
									<span>Agent type is required</span>
								</div>
							)}
						</div>

						<Button
							disabled={!formData.name || !formData.platform || !formData.agentType}
							className="bg-blue-500 w-full hover:bg-blue-600">
							Generate Config
						</Button>
						{showRunCommand && (
							<div className="mt-2 flex flex-col gap-2 mb-4">
								<p className="text-lg font-bold text-black">Run Command</p>
								<p className="text-gray-500">
									Running this command in your selected envoirment will deploy the {formData.agentType === "fluent-bit" ? "Fluent Bit" : "OpenTelemetry"} pipeline
								</p>
								<div className="flex justify-between border-2 border-orange-300 p-3 rounded-lg text-orange-400">
									<p>
										{formData.platform
											? (formData.agentType === "fluent-bit"
												? installCommandsFluentBit[formData.platform as keyof typeof installCommandsFluentBit](formData.name)
												: installCommands[formData.platform as keyof typeof installCommands](formData.name))
											: "Select a platform to see the command"}
									</p>
									{formData.platform && (
										<CopyIcon
											onClick={() =>
												handleCopy(
													formData.agentType === "fluent-bit"
														? installCommandsFluentBit[formData.platform as keyof typeof installCommandsFluentBit](formData.name)
														: installCommands[formData.platform as keyof typeof installCommands](formData.name),
												)
											}
											className="h-8 w-8 text-orange-400 cursor-pointer"
										/>
									)}
								</div>
							</div>
						)}
						{showHeartBeat && (
							<div className="mt-3 flex flex-col gap-2">
								{/* <p>Once the agent is completely installed it will also appear in the Agent list Table</p> */}
								<div className="flex gap-4 border-2 border-blue-300 p-3 rounded-lg text-blue-400">
									<Loader2 className="h-5 w-5 text-blue-400 animate-spin" />
									<p>CtrlB is checking for heartbeat..</p>
								</div>
							</div>
						)}

						{status === "success" && showStatus ? (
							<div className="mt-3 bg-green-200 flex p-3 gap-2 items-center rounded-md">
								<BadgeCheck className="text-green-600" />
								<p className="text-green-600">Your agent is successfully deployed</p>
							</div>
						) : showStatus && !showHeartBeat ? (
							<div className="mt-3 bg-red-200 flex p-3 gap-2 items-center justify-between rounded-md">
								<div className="flex justify-start items-center gap-1">
									<AlertCircle className="text-red-600 h-5 w-5" />
									<p className="text-red-600">Heartbeat not detected</p>
								</div>
								<Button variant={"destructive"} onClick={handleTryAgain}>
									Try again
								</Button>
							</div>
						) : null}
					</form>
					{showConfigureButton && (
						<div className="flex justify-end mt-3">
							<Button
								onClick={() => {
									try {
										// First, clear any potentially corrupted data
										const keysToRemove = [
											"pipelineData",
											"latest_agents",
											"selectedAgentIds",
											"pipelineNodes",
											"pipelineEdges",
										];
										keysToRemove.forEach(key => localStorage.removeItem(key));

										// Initialize fresh pipeline data
										const initialPipelineData = {
											id: Date.now().toString(),
											name: formData.name,
											platform: formData.platform,
											agentType: formData.agentType,
											nodes: [],
											edges: [],
											created_at: new Date().toISOString(),
										};

										// Store all required data with proper JSON formatting
										localStorage.setItem("pipelinename", formData.name);
										localStorage.setItem("platform", formData.platform);
										localStorage.setItem("agentType", formData.agentType);
										localStorage.setItem("pipelineData", JSON.stringify(initialPipelineData));

										// Verify data was stored correctly
										const verifyData = localStorage.getItem("pipelineData");
										if (!verifyData) {
											throw new Error("Failed to store pipeline data");
										}

										// Move to next step
										setCurrentStep(currentStep + 1);
									} catch (error) {
										console.error("Error initializing pipeline:", error);
										showSnackbar(
											"Failed to initialize pipeline data. Please try again.",
											"error",
											3000
										);
									}
								}}
								// disabled={!formData.name || !formData.platform || !EDI_API_KEY}
								disabled={!formData.name || !formData.platform || !formData.agentType}
								className={`px-6 ${
									status === "success" ? "bg-blue-500 hover:bg-blue-600" : "bg-gray-400 hover:bg-gray-500"
								}`}>
								Configure Pipeline
							</Button>
						</div>
					)}
				</CardContent>
			</Card>
		</div>
	);
};

export default AddPipelineDetails;
