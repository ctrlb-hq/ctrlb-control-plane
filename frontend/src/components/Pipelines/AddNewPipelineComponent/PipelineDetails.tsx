import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label';
import { usePipelineStatus } from '@/context/usePipelineStatus';
import { AlertCircle, Badge, CopyIcon, Loader2 } from 'lucide-react';
import { useState } from 'react';
import ProgressFlow from './ProgressFlow';

import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem, SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import { useToast } from '@/hooks/use-toast';
import { Close } from '@radix-ui/react-dialog';
import agentServices from '@/services/agentServices';

interface formData {
    name: string,
    platform: string
}

const PipelineDetails = () => {
    const pipelineStatus = usePipelineStatus();
    if (!pipelineStatus) {
        return null;
    }
    const pipelineName = localStorage.getItem("pipelinename") || '';
    const platform = localStorage.getItem("platform")

    const { currentStep } = pipelineStatus;
    const [showRunCommand, setShowRunCommand] = useState(false)
    const [showHeartBeat, setShowHeartBeat] = useState(false)
    const [showStatus, setShowStatus] = useState(false)
    const [status, setStatus] = useState<"success" | "failed">("failed")
    const [showAgentInfo, setShowAgentInfo] = useState(false)
    const { toast } = useToast()
    const EDI_API_KEY = "b684f7-9485ght-4f7-9f8g-4f7g9-4f7g9"


    const [formData, setFormData] = useState<formData>({
        name: pipelineName ?? '',
        platform: platform ?? "",
    });

    const [errors, setErrors] = useState({
        name: false,
        platform: false,
    });

    const [touched, setTouched] = useState({
        name: false,
        platform: false,
    });

    const handleChange = (e: any) => {
        const { id, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [id]: value
        }));

        // Clear error when user types
        if (value.trim()) {
            setErrors(prev => ({
                ...prev,
                [id]: false
            }));
        }
    };

    const handleSubmit = (e: any) => {
        e.preventDefault();
        // Check required fields
        const newErrors = {
            name: !formData.name.trim(),
            platform: !formData.platform
        };

        setErrors(newErrors);
        setTouched({
            name: true,
            platform: true
        });

        setShowRunCommand(true)

    };


    const handleCopy = () => {
        navigator.clipboard.writeText(`${EDI_API_KEY}`)
        const since = new Date().getTime()
        setTimeout(() => {
            toast({
                title: 'Copied',
                description: 'API Key copied to clipboard',
                duration: 2000,
            })
        }, 1000)
        setTimeout(() => {
            setShowHeartBeat(true)
        }, 2000)
        setTimeout(() => {
            setShowStatus(true)
        }, 6000)
        setTimeout(() => {
            setShowAgentInfo(true)
            checkAgentStatus(since)
        }, 1000)
    }


    const checkAgentStatus = async (since: number) => {
        try {
            const interval = setInterval(async () => {
                const agents = await agentServices.getLatestAgents({since});
                console.log("first",agents)
                if (agents && agents.length > 0) {
                    setStatus("success");
                    setShowStatus(true);
                    clearInterval(interval);
                }
            }, 2*1000); // Check every 2 seconds

            // Clear interval after 30 seconds (timeout)
            setTimeout(() => {
                clearInterval(interval);
                if (status !== "success") {
                    setStatus("failed");
                    setShowStatus(true);
                }
            }, 30000);
        } catch (error) {
            console.error("Error checking agent status:", error);
            setStatus("failed");
            setShowStatus(true);
        }
    };


    return (
        <div className='flex flex-col gap-5'>
            <div className=" flex gap-5">
                <div className='flex w-[30rem]'>
                <ProgressFlow />
                </div>
                <Card className="w-[39rem] h-[45rem]">
                    <CardHeader>
                        <CardTitle className="text-xl font-bold">
                            Let's get started building your Pipeline.
                        </CardTitle>

                        <p className="text-gray-600 mt-2">
                            Let's get started building your pipeline configuration.
                        </p>
                    </CardHeader>
                    <CardContent className='h-[27rem]'>
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
                                    className={`h-10 ${errors.name && touched.name ? 'border-red-500 focus-visible:ring-red-500' : 'border-gray-300'}`}
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
                                <Select
                                    value={formData.platform}
                                    onValueChange={(value: string) => {
                                        setFormData(prev => ({
                                            ...prev,
                                            platform: value
                                        }));

                                        // Clear error when user selects
                                        if (value.length > 0) {
                                            setErrors(prev => ({
                                                ...prev,
                                                platform: false
                                            }));
                                        }
                                    }}
                                    required
                                >
                                    <SelectTrigger className={`h-10 w-full border rounded-md px-3 py-2 ${errors.platform && touched.platform ? 'border-red-500 focus-visible:ring-red-500' : 'border-gray-300'}`}>
                                        <SelectValue placeholder="Select a platform" />
                                    </SelectTrigger>

                                    <SelectContent>
                                        <SelectGroup>
                                            <SelectItem value="linux">Linux</SelectItem>
                                            <SelectItem value="kubernetes">Kubernetes</SelectItem>
                                            <SelectItem value="macOS">macOS</SelectItem>
                                            <SelectItem value="openShift">openShift</SelectItem>
                                        </SelectGroup>
                                    </SelectContent>
                                </Select>
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
                            <Button disabled={!formData.name || !formData.platform} className='bg-blue-500 w-full hover:bg-blue-600'>
                                Generate Config
                            </Button>
                            {showRunCommand && <div className="mt-2 flex flex-col gap-2 mb-4">
                                <p className="text-lg font-bold text-black">Run Command</p>
                                <p className="text-gray-500">Running this command in your selected envoirment will deploy the pipeline</p>
                                <div className="flex justify-between border-2 border-orange-300 p-3 rounded-lg text-orange-400">
                                    <p>EDI_API_KEY={EDI_API_KEY}</p>
                                    <CopyIcon onClick={handleCopy} className="h-5 w-5 text-orange-400 cursor-pointer" />
                                </div>
                            </div>}
                            {showHeartBeat && <div className="mt-3 flex flex-col gap-2">
                                <p>Once the agent is completely installed it will also appear in the Agent list Table</p>
                                <div className="flex gap-4 border-2 border-blue-300 p-3 rounded-lg text-blue-400">
                                    <Loader2 className="h-5 w-5 text-blue-400 animate-spin" />
                                    <p>CtrlB is checking for heartbeat..</p>
                                </div>
                            </div>}
                            {status === "success" ? showStatus && <div className="mt-3 bg-green-200 flex p-3 gap-2 items-center rounded-md">
                                <Badge className="text-green-600" />
                                <p className="text-green-600">Your agent is successfully deployed</p>
                            </div> : showStatus && <div className="mt-3 bg-red-200 flex p-3 gap-2 items-center justify-between rounded-md">
                                <div className="flex justify-start">
                                    <Close className="text-red-600" />
                                    <p className="text-red-600">Heartbeat not detected</p>
                                </div>
                                <Button variant={"destructive"}>Try again</Button>
                            </div>}
                        </form>
                        <div className='flex justify-end mt-3'>
                            <Button
                                onClick={() => {
                                    localStorage.setItem('pipelinename', formData.name)
                                    localStorage.setItem('platform', formData.platform)
                                    pipelineStatus.setCurrentStep(currentStep + 1);
                                    handleSubmit
                                }}
                                disabled={!formData.name || !formData.platform}
                                className='bg-blue-500 px-6 hover:bg-blue-600'>
                                Configure Pipeline
                            </Button>
                        </div>
                    </CardContent>

                </Card>
            </div>
        </div>


    )
}

export default PipelineDetails