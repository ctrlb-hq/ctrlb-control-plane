import { PencilIcon } from "@heroicons/react/24/solid";
import { Typography, Chip, Avatar, Tooltip } from "@material-tailwind/react";
import { useNavigate } from "react-router-dom";


const TABLE_HEAD = ["Name", "Type", "Status", "Exported Volume", ""];

const TABLE_ROWS = [
    {
        img: "https://cdn.brandfetch.io/idxVhszl6V/w/400/h/400/theme/dark/icon.jpeg-1.jpg",
        name: "Agent Alpha",
        type: "Linux",
        version: "v2.0.1",
        status: "Active",
        exportedVolume: "150 GB",
    },
    {
        img: "https://cdn.brandfetch.io/idxVhszl6V/w/400/h/400/theme/dark/icon.jpeg",
        name: "Agent Beta",
        type: "Windows",
        version: "v1.3.5",
        status: "Inactive",
        exportedVolume: "85 GB",
    },
];


export function AgentsTable() {
    const navigate = useNavigate();
    const handleClick = () => navigate("/config/123");

    return (
        <div className="bg-white rounded-lg overflow-hidden flex-grow border border-gray-200">
            <div>
                <table className="w-full table-fixed">
                    <thead className="bg-gray-50">
                        <tr>
                            {TABLE_HEAD.map((head) => (
                                <th key={head} className="py-3 px-4 text-left font-semibold text-sm text-gray-600">
                                    {head}
                                </th>
                            ))}
                        </tr>
                    </thead>
                </table>
            </div>
            <div className="overflow-y-auto max-h-[calc(100vh-200px)]">
                <table className="w-full table-fixed">
                    <tbody>
                        {TABLE_ROWS.map(({ img, name, type, version, status, exportedVolume }) => (
                            <tr key={name} className="border-b border-gray-200 hover:bg-gray-50">
                                <td className="py-3 px-4">
                                    <div className="flex items-center gap-3">
                                        <Avatar
                                            src={img}
                                            alt={name}
                                            className="h-10 w-10"
                                            placeholder={undefined}
                                            onPointerEnterCapture={undefined}
                                            onPointerLeaveCapture={undefined} />
                                        <Typography
                                            className="text-sm text-gray-800 font-medium"
                                            placeholder={undefined}
                                            onPointerEnterCapture={undefined}
                                            onPointerLeaveCapture={undefined}
                                        >
                                            {name}
                                        </Typography>
                                    </div>
                                </td>
                                <td className="py-3 px-4">
                                    <Typography
                                        className="text-sm text-gray-800"
                                        placeholder={undefined}
                                        onPointerEnterCapture={undefined}
                                        onPointerLeaveCapture={undefined}
                                    >
                                        {type}
                                    </Typography>
                                    <Typography
                                        className="text-xs text-gray-500"
                                        placeholder={undefined}
                                        onPointerEnterCapture={undefined}
                                        onPointerLeaveCapture={undefined}
                                    >
                                        {version}
                                    </Typography>
                                </td>
                                <td className="py-3 px-4">
                                    <Chip
                                        size="sm"
                                        value={status}
                                        className={`${status === 'Active' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'
                                            } px-2 py-1 text-xs font-medium rounded-full w-20 text-center`}
                                    />
                                </td>
                                <td className="py-3 px-4">
                                    <Typography
                                        className="text-sm text-gray-800"
                                        placeholder={undefined}
                                        onPointerEnterCapture={undefined}
                                        onPointerLeaveCapture={undefined}
                                    >
                                        {exportedVolume}
                                    </Typography>
                                </td>
                                <td className="py-3 px-4">
                                    <Tooltip content="Edit Config">
                                        <PencilIcon onClick={handleClick} className="h-5 w-5 text-gray-600 cursor-pointer" />
                                    </Tooltip>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
}