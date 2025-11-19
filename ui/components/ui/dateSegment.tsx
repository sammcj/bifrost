"use client";

import { RefObject, useRef } from "react";
import { useDateSegment } from "react-aria";
import { DateFieldState, DateSegment as IDateSegment } from "react-stately";
import { cn } from "@/lib/utils";

interface DateSegmentProps {
	segment: IDateSegment;
	state: DateFieldState;
}

function DateSegment({ segment, state }: DateSegmentProps) {
	const ref = useRef<HTMLDivElement>(null);

	const {
		segmentProps: { ...segmentProps },
	} = useDateSegment(segment, state, ref as RefObject<HTMLElement>);

	return (
		<div
			{...segmentProps}
			ref={ref}
			className={cn(
				"focus:bg-background-highlight-primary focus:bg-blue-400 dark:focus:bg-blue-500 focus:text-content-primary focus:rounded-[2px] focus:outline-hidden",
				segment.type !== "literal" ? "px-[1px]" : "",
				segment.isPlaceholder ? "text-content-disabled" : "",
			)}
		>
			{segment.text}
		</div>
	);
}

export { DateSegment };
