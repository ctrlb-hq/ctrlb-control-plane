import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

type Props = {
	name: string;
	isEditMode: boolean;
	setIsEditMode: (v: boolean) => void;
	onReviewClick: () => void;
	onClose: () => void;
	isDeploying: boolean;
};

const PipelineEditorHeader = ({
	name,
	isEditMode,
	setIsEditMode,
	onReviewClick,
	onClose,
	isDeploying,
}: Props) => {
	return (
		<div className="flex justify-between items-center p-4 border-b">
			<div className="text-xl font-medium">{name}</div>

			<div className="flex items-center mr-6 space-x-4">
				<Switch
					id="edit-mode"
					checked={isEditMode}
					onCheckedChange={setIsEditMode}
				/>
				<Label htmlFor="edit-mode">Edit Mode</Label>

				<Button
					disabled={!isEditMode}
					onClick={onReviewClick}
					className={
						isEditMode
							? "bg-blue-500 hover:bg-blue-600 text-white"
							: "bg-blue-200 text-blue-500 cursor-not-allowed"
					}
				>
					Review
				</Button>

				<Button
					variant="destructive"
					onClick={onClose}
					disabled={isDeploying}
					className={
						isDeploying
							? "bg-red-200 text-red-500 cursor-not-allowed hover:bg-red-200"
							: "bg-red-500 hover:bg-red-600 text-white"
					}
				>
					Close
				</Button>
			</div>
		</div>
	);
};

export default PipelineEditorHeader;
