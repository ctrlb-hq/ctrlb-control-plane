import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import { PlusIcon, Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { usePipelineStatus } from "@/context/usePipelineStatus";
import ProgressFlow from "../ProgressFlow";
import {
    Sheet,
    SheetClose,
    SheetContent,
    SheetTrigger,
} from "@/components/ui/sheet";
import Tabs from "../Tabs";
import { usePipelineTab } from "@/context/useAddNewPipelineActiveTab";
import CreateNewAgent from "@/components/Agents/CreateNewAgent";
import { TransporterService } from "@/services/transporterService";

interface sources {
    name: string,
    display_name: string,
    type: string,
    supported_signals: string[]
}

import { JsonForms } from '@jsonforms/react';

import {
    materialCells,
    materialRenderers,
} from '@jsonforms/material-renderers';

import { ThemeProvider, createTheme } from '@mui/material/styles';

const theme = createTheme({
    components: {
        MuiFormControl: {
            styleOverrides: {
                root: {
                    marginBottom: '0.5rem',
                },
            },
        },
    },
});

const renderers = [
    ...materialRenderers,
];
const DestinationDetail = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedSource, setSelectedSource] = useState<sources | null>(null);
    const [showSourceSheet, setShowSourceSheet] = useState(false);
    const [editSourceSheet, setEditSourceSheet] = useState(false)
    const { currentTab } = usePipelineTab()
    const [sources, setSources] = useState<sources[]>([])
    const [submitDisabled, setSubmitDisabled] = useState(true);

    const [form, setForm] = useState<object>({})
    const [nodes, setNodes] = useState<object[]>([])
    const [existingSources, setExistingSources] = useState<sources[]>(() => {
        const savedSources = localStorage.getItem(`Destination`);
        return savedSources ? JSON.parse(savedSources) : [];
    });

    const [data, setData] = useState<object>();

    const source = async () => {
        const res = await TransporterService.getTransporterService("exporter")
        setSources(res)
    }
    const getForm = async () => {
        const res = await TransporterService.getTransporterForm(selectedSource!.name)
        setForm(res)

    }
    const filteredSources = sources.filter(source =>
        source.name.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const pipelineStatus = usePipelineStatus();
    if (!pipelineStatus) {
        return null;
    }
    useEffect(() => {
        getForm()
        source()
    }, [selectedSource])


    let { currentStep, setCurrentStep } = pipelineStatus;

    const handleDeleteSource = (index: number) => {
        const updatedSources = existingSources.filter((_, i) => i !== index);
        setExistingSources(updatedSources);
        localStorage.setItem(`Destination`, JSON.stringify(updatedSources));

        const destinationToDelete = existingSources[index];
        const existingNodes = JSON.parse(localStorage.getItem('Nodes') || '[]');
        const updatedNodes = existingNodes.filter((node: any) => node.component_name !== destinationToDelete.name);
        localStorage.setItem('Nodes', JSON.stringify(updatedNodes));
    };


    const handleSourceConfiguration = (source: sources) => {
        setSelectedSource(source);
    }

    const handleSubmit = () => {
        const log={ type: 'destination', name: selectedSource?.display_name, status: "added" }
        const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
        const updatedLog = [...existingLog, log];
        const updatedSources = [
            ...existingSources,
            {
                name: selectedSource!.name,
                display_name: selectedSource!.display_name,
                supported_signals: selectedSource!.supported_signals,
                type: selectedSource!.type,
            },
        ];
        const existingNodes = JSON.parse(localStorage.getItem('Nodes') || '[]');

        const updatedNodes = [
            ...nodes,
            {
                component_id: existingNodes.length + 1,
                name: selectedSource!.display_name,
                component_role: selectedSource!.type,
                component_name: selectedSource!.name,
                config: data,
                supported_signals: selectedSource!.supported_signals
            }
        ];
        setNodes(updatedNodes);
        setExistingSources(updatedSources);
        localStorage.setItem(`Destination`, JSON.stringify(updatedSources));
        const newNodes = [...existingNodes.filter(node => !updatedNodes.some(updatedNode => updatedNode.component_id === node.component_id)), ...updatedNodes];
        localStorage.setItem(`Nodes`, JSON.stringify(newNodes));
        localStorage.setItem("changesLog", JSON.stringify(updatedLog));

        

    };

    const handleEdit = () => {
        if (editSourceSheet) {
            const existingNodes = JSON.parse(localStorage.getItem('Nodes') || '[]');
            const existingSources = JSON.parse(localStorage.getItem('Destination') || '[]');

            const isDuplicateNode = existingNodes.some((node: any) =>
                node.component_name === selectedSource!.name && JSON.stringify(node.config) === JSON.stringify(data)
            );

            const isDuplicateSource = existingSources.some((source: any) =>
                source.name === selectedSource!.name
            );

            if (!isDuplicateNode || !isDuplicateSource) {
                const updatedSources = [
                    ...existingSources,
                    {
                        name: selectedSource!.name,
                        display_name: selectedSource!.display_name,
                        supported_signals: selectedSource!.supported_signals,
                        type: "Destination",
                    },
                ];

                const updatedNodes = [
                    ...existingNodes,
                    {
                        component_id: existingNodes.length + 1,
                        name: selectedSource!.display_name,
                        component_role: selectedSource!.type,
                        component_name: selectedSource!.name,
                        config: data,
                        supported_signals: selectedSource!.supported_signals
                    }
                ];

                localStorage.setItem('Destination', JSON.stringify(updatedSources));
                localStorage.setItem('Nodes', JSON.stringify(updatedNodes));
                setExistingSources(updatedSources);
                setEditSourceSheet(false)
            }
        }
    };


    return (
        <div className="flex flex-col gap-5">
            <Tabs />
            {currentTab == "pipelines" ? <div className="mx-auto flex gap-5 w-full">
                <ProgressFlow />
                <Card className="w-full h-[40rem] bg-white shadow-sm">
                    <CardContent className="p-6 h-[36rem]">
                        <div className="space-y-4">
                            <h2 className="text-xl font-semibold text-gray-700">
                                Add Destination from which you'd like to collect telemetry.
                            </h2>

                            <p className="text-gray-600 text-sm">
                                A Destination is a combination of OpenTelemetry receivers and
                                processors that allows you to collect telemetry from a specific
                                technology. Ensuring the right combination of these components is
                                one of the most challenging aspects of building an OpenTelemetry
                                configuration file. With CtrlB, we handle that all for you.
                            </p>

                            {existingSources && (
                                existingSources
                                    .filter((source: sources) => source.name && source.display_name)
                                    .map((source: sources, index: number) => (
                                        <div
                                            className="flex justify-between items-center border rounded-md border-gray-300 p-3"
                                            key={index}
                                        >
                                            <div className="capitalize">
                                                {source.type} | {source.display_name}
                                            </div>
                                            <div className="flex gap-2">
                                                <Sheet
                                                    open={editSourceSheet}
                                                    onOpenChange={(open) => setEditSourceSheet(open)}
                                                >
                                                    <SheetTrigger asChild>
                                                        <Button
                                                            variant={"outline"}
                                                            onClick={() => {
                                                                const nodes = JSON.parse(localStorage.getItem('Nodes') || '[]');
                                                                const node = nodes.find((n: any) => n.component_name === source.name);
                                                                if (node) {
                                                                    setData(node.config);
                                                                    setSelectedSource(source);
                                                                }
                                                            }}
                                                        >
                                                            Edit
                                                        </Button>
                                                    </SheetTrigger>
                                                    <SheetContent>
                                                        <ThemeProvider theme={theme}>
                                                            <div className='mt-3'>
                                                                <div className='text-2xl p-4 font-semibold bg-gray-100'>{form.title}</div>
                                                                <div className='p-3 '>
                                                                    <div className='overflow-y-auto h-[45rem]'>
                                                                        <JsonForms
                                                                            data={data}
                                                                            schema={form}
                                                                            renderers={renderers}
                                                                            cells={materialCells}
                                                                            onChange={({ data }) => setData(data)}
                                                                        />
                                                                        <SheetClose>
                                                                            <div className='flex justify-end mb-10'>
                                                                                <Button size={"lg"} className='bg-blue-500' onClick={() => {
                                                                                    handleEdit();
                                                                                    setSelectedSource(null);
                                                                                }}>
                                                                                    Submit
                                                                                </Button>
                                                                            </div>
                                                                        </SheetClose>
                                                                    </div>
                                                                </div>
                                                            </div>
                                                        </ThemeProvider>
                                                    </SheetContent>
                                                </Sheet>
                                                <Button
                                                    variant={"destructive"}
                                                    onClick={() => handleDeleteSource(index)}
                                                >
                                                    Delete
                                                </Button>
                                            </div>
                                        </div>
                                    ))
                            )}

                            <Sheet open={showSourceSheet} onOpenChange={(open) => setShowSourceSheet(open)}>
                                <SheetTrigger asChild>
                                    <Button className="flex items-center w-full gap-1 px-4 py-1 bg-blue-500 text-white" variant="outline">Add Destination
                                        <PlusIcon className="h-4 w-4" />
                                    </Button>
                                </SheetTrigger>
                                <SheetContent>
                                    <div className="p-4">
                                        <p className="text-2xl mb-3">Add Destination</p>
                                        <div className="relative">
                                            <Search className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
                                            <Input
                                                placeholder="Search for a technology..."
                                                className="pl-10 pr-4 py-2 border rounded-md w-full"
                                                value={searchTerm}
                                                onChange={(e) => setSearchTerm(e.target.value)}
                                            />
                                        </div>
                                        <div className="flex-1 overflow-auto">
                                            <div className="p-4 h-[40rem]">
                                                {filteredSources.map((source: sources) => (
                                                    <Sheet >
                                                        <SheetTrigger asChild>
                                                            <div onClick={() => handleSourceConfiguration(source)} className="flex items-center justify-between p-3 hover:bg-gray-50 border-b cursor-pointer">
                                                                <div className="flex items-center">
                                                                    <span className="ml-3 font-medium">{source.display_name}</span>
                                                                </div>
                                                                <div className="flex space-x-1">
                                                                    {source.supported_signals.map((feature: string) => (
                                                                        <span key={feature} className="px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-md">
                                                                            {feature}
                                                                        </span>
                                                                    ))}
                                                                </div>
                                                            </div>
                                                        </SheetTrigger>
                                                        <SheetContent>
                                                            <ThemeProvider theme={theme}>
                                                                <div className='mt-3'>
                                                                    <div className='text-2xl p-4 font-semibold bg-gray-100'>{form.title}</div>
                                                                    <div className='p-3 '>
                                                                        <div className='overflow-y-auto h-[45rem]'>
                                                                            <JsonForms
                                                                                data={data}
                                                                                schema={form}
                                                                                renderers={renderers}
                                                                                cells={materialCells}
                                                                                onChange={({ data, errors }) => {
                                                                                    setData(data);
                                                                                    const hasErrors = errors && errors.length > 0;
                                                                                    setSubmitDisabled(!!hasErrors);
                                                                                }}                                                                            />
                                                                            <SheetClose>
                                                                                <div className='flex justify-end mb-10'>
                                                                                    <Button size={"lg"} className='bg-blue-500' onClick={() => {
                                                                                        handleSubmit();
                                                                                        setSelectedSource(null);
                                                                                        setShowSourceSheet(false);
                                                                                        
                                                                                    }}
                                                                                    disabled={submitDisabled}>
                                                                                        Submit
                                                                                    </Button>
                                                                                </div>
                                                                            </SheetClose>
                                                                        </div>
                                                                    </div>
                                                                </div>
                                                            </ThemeProvider>
                                                        </SheetContent>
                                                    </Sheet>
                                                ))}
                                            </div>
                                        </div>
                                    </div>
                                </SheetContent>
                            </Sheet>
                        </div>
                        {/* {selectedSource && (
                            <Sheet>
                                <SheetTrigger asChild>
                                    <Button className="flex items-center gap-1 px-4 py-1 bg-blue-500 text-white" variant="outline">Add New Pipeline
                                        <PlusIcon className="h-4 w-4" />
                                    </Button>
                                </SheetTrigger>
                                <SheetContent>
                                </SheetContent>
                            </Sheet>
                        )} */}
                    </CardContent>
                    <CardFooter className="flex justify-end items-end">
                        <div className=" flex items-end justify-end gap-4">
                            <Button
                                className="bg-gray-700 px-6 disabled:opacity-50"
                                disabled={currentStep === 0}
                                onClick={() => setCurrentStep(--currentStep)}
                            >
                                Back
                            </Button>
                            <Button
                                className="bg-blue-500 hover:bg-blue-700 px-6 disabled:opacity-50"
                                onClick={() => setCurrentStep(++currentStep)}
                            >
                                Next
                            </Button>
                        </div>
                    </CardFooter>
                </Card>
            </div> : <CreateNewAgent />}

        </div>
    )
}

export default DestinationDetail;