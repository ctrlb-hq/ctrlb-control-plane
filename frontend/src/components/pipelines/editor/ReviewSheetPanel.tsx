import { Drawer, Box, Typography, Button, IconButton, Divider } from "@mui/material";
import { Edit, Loader2 } from "lucide-react";

type Props = {
	open: boolean;
	onOpenChange: (v: boolean) => void;
	isReviewOpen: boolean;
	changesLog: any[];
	isDeploying: boolean;
	onDeploy: () => void;
	onEdit: (change: any) => void;
};

const ReviewSheetPanel = ({
	open,
	onOpenChange,
	isReviewOpen,
	changesLog,
	isDeploying,
	onDeploy,
	onEdit,
}: Props) => {
	return (
		<Drawer
			anchor="right"
			open={open}
			onClose={() => onOpenChange(false)}
			disableEnforceFocus
			ModalProps={{
				keepMounted: true,
			}}
			sx={{
				'& .MuiDrawer-paper': {
					width: '30rem',
					padding: '1.5rem',
				},
			}}
		>
			<Box sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
				{isReviewOpen && (
					<>
						<Typography variant="h6" component="h2" sx={{ fontWeight: 600, mb: 1 }}>
							Pending Changes
						</Typography>
						
						<Box
							sx={{
								display: 'flex',
								flexDirection: 'column',
								gap: 3,
								mt: 2,
								overflow: 'auto',
								height: '40rem',
								pr: 1,
							}}
						>
							{changesLog.map((change, index) => (
								<Box key={index}>
									<Box
										sx={{
											display: 'flex',
											justifyContent: 'space-between',
											alignItems: 'center',
										}}
									>
										<Box>
											<Typography variant="body1" sx={{ fontSize: '1.125rem' }}>
												{change.type}
											</Typography>
											<Typography
												variant="body2"
												sx={{ color: 'rgb(31, 41, 55)' }}
											>
												{change.name}
											</Typography>
										</Box>
										<Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5 }}>
											<Typography
												variant="body1"
												sx={{
													fontSize: '1.125rem',
													color: 'rgb(107, 114, 128)',
												}}
											>
												[{change.status}]
											</Typography>
											{change.type !== "Edge" && (
												<IconButton
													onClick={() => onEdit(change)}
													size="small"
													sx={{ padding: 0.5 }}
												>
													<Edit className="w-6 h-6 cursor-pointer" />
												</IconButton>
											)}
										</Box>
									</Box>
									{index < changesLog.length - 1 && (
										<Divider sx={{ mt: 2 }} />
									)}
								</Box>
							))}
						</Box>

						<Button
							onClick={onDeploy}
							disabled={isDeploying}
							variant="contained"
							sx={{
								mt: 2,
								bgcolor: 'rgb(59, 130, 246)',
								'&:hover': {
									bgcolor: 'rgb(37, 99, 235)',
								},
								textTransform: 'none',
								display: 'flex',
								gap: 1,
							}}
						>
							{isDeploying ? (
								<>
									<Loader2 className="animate-spin h-4 w-4" />
									Deploying…
								</>
							) : (
								"Deploy Changes"
							)}
						</Button>
					</>
				)}
			</Box>
		</Drawer>
	);
};

export default ReviewSheetPanel;