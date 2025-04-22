import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label';
import { usePipelineStatus } from '@/context/usePipelineStatus';
import { AlertCircle, Badge, CopyIcon, Loader2 } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
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
import { useNavigate } from 'react-router-dom';

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
    const [isApiKeyCopied, setIsApiKeyCopied] = useState(false);
    const [showConfigureButton, setShowConfigureButton] = useState(false);
    const EDI_API_KEY = "b684f7-9485ght-4f7-9f8g-4f7g9-4f7g9"

    const [isChecking, setIsChecking] = useState(false);
    const abortControllerRef = useRef<AbortController | null>(null);
    const navigate = useNavigate();


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
        setIsApiKeyCopied(true); 
        setShowConfigureButton(true); 
        const since = Math.floor(new Date().getTime() / 1000)
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

    const handleTryAgain = () => {
        setShowStatus(false);
        setStatus("failed");
        setShowHeartBeat(false);

        setTimeout(() => {
            const since = Math.floor(new Date().getTime() / 1000)
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

    // Cleanup on unmount
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

                    const agents = await agentServices.getLatestAgents({ since });
                    if (agents) {
                        setStatus(agents ? "success" : "failed");
                        setShowStatus(true);
                        setShowHeartBeat(false);
                        stopChecking();
                        if (agents) {
                            navigate('/PipelineCanvas'); 
                        }
                        break;
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

                            {status === "success" && showStatus ? (
                                <div className="mt-3 bg-green-200 flex p-3 gap-2 items-center rounded-md">
                                    <Badge className="text-green-600" />
                                    <p className="text-green-600">Your agent is successfully deployed</p>
                                </div>
                            ) : (showStatus && !showHeartBeat) ? (
                                <div className="mt-3 bg-red-200 flex p-3 gap-2 items-center justify-between rounded-md">
                                    <div className="flex justify-start">
                                        <Close className="text-red-600" />
                                        <p className="text-red-600">Heartbeat not detected</p>
                                    </div>
                                    <Button variant={"destructive"} onClick={handleTryAgain}>
                                        Try again
                                    </Button>
                                </div>
                            ) : null}
                        </form>
                        {showConfigureButton && (
                            <div className='flex justify-end mt-3'>
                                <Button
                                    onClick={() => {
                                        localStorage.setItem('pipelinename', formData.name)
                                        localStorage.setItem('platform', formData.platform)
                                        pipelineStatus.setCurrentStep(currentStep + 1);
                                        handleSubmit
                                    }}
                                    disabled={!formData.name || !formData.platform || !EDI_API_KEY}
                                    className='bg-blue-500 px-6 hover:bg-blue-600'>
                                    Configure Pipeline
                                </Button>
                            </div>
                        )}
                    </CardContent>

                </Card>
            </div>
        </div>


    )
}

export default PipelineDetails