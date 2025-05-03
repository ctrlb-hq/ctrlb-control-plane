import React, { useState } from "react";
import { X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Command,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
} from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

interface MultiSelectProps {
	options: { value: string; label: string }[];
	value?: string[];
	onChange: (value: string[]) => void;
	placeholder?: string;
}

export function MultiSelect({
	options,
	value = [],
	onChange,
	placeholder = "Select items",
}: MultiSelectProps) {
	const [open, setOpen] = useState(false);

	const handleSelect = (option: string) => {
		const newValue = value.includes(option) ? value.filter(v => v !== option) : [...value, option];

		onChange(newValue);
	};

	const handleRemove = (option: string) => {
		onChange(value.filter(v => v !== option));
	};

	return (
		<Popover open={open} onOpenChange={setOpen}>
			<PopoverTrigger asChild>
				<Button
					variant="outline"
					role="combobox"
					aria-expanded={open}
					className="w-full justify-between"
				>
					<div className="flex flex-wrap gap-1">
						{value.length === 0 ? (
							<span className="text-muted-foreground">{placeholder}</span>
						) : (
							options
								.filter(option => value.includes(option.value))
								.map(option => (
									<Badge key={option.value} variant="secondary" className="flex items-center">
										{option.label}
										<X
											className="ml-1 h-3 w-3 cursor-pointer"
											onClick={e => {
												e.stopPropagation();
												handleRemove(option.value);
											}}
										/>
									</Badge>
								))
						)}
					</div>
				</Button>
			</PopoverTrigger>
			<PopoverContent className="w-full p-0">
				<Command>
					<CommandInput placeholder="Search items..." />
					<CommandList>
						<CommandEmpty>No items found.</CommandEmpty>
						<CommandGroup>
							{options.map(option => (
								<CommandItem
									key={option.value}
									value={option.value}
									onSelect={() => handleSelect(option.value)}
								>
									<div className="flex items-center">
										<div
											className={`mr-2 h-4 w-4 border ${
												value.includes(option.value) ? "bg-primary text-primary-foreground" : "bg-background"
											}`}
										>
											{value.includes(option.value) && "✓"}
										</div>
										{option.label}
									</div>
								</CommandItem>
							))}
						</CommandGroup>
					</CommandList>
				</Command>
			</PopoverContent>
		</Popover>
	);
}
