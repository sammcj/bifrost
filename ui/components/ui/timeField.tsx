import { Time } from "@internationalized/date";
import { RefObject, useRef } from "react";
import { AriaTimeFieldProps, TimeValue, useLocale, useTimeField } from "react-aria";
import { useTimeFieldState } from "react-stately";
import { cn } from "@/lib/utils";
import { DateSegment } from "@/components/ui/dateSegment";

interface TimeFieldProps extends AriaTimeFieldProps<TimeValue> {
	className?: string;
	hourCycle?: 12 | 24;
}

function TimeField(props: TimeFieldProps) {
	const ref = useRef<HTMLDivElement | null>(null);

	const { locale } = useLocale();

	// Set proper placeholder value for 12-hour format to help with "12" input handling
	const defaultPlaceholderValue = new Time(12, 0); // 12:00 PM as placeholder for better 12-hour handling

	const state = useTimeFieldState({
		...props,
		locale,
		hourCycle: props.hourCycle ?? 12, // Default to 12-hour format
		placeholderValue: props.placeholderValue ?? defaultPlaceholderValue, // Use 12:00 as default placeholder
		// Force leading zeros to help with input parsing
		shouldForceLeadingZeros: true,
	});

	const {
		fieldProps: { ...fieldProps },
		labelProps,
	} = useTimeField(props, state, ref as RefObject<Element>);

	return (
		<div
			{...fieldProps}
			ref={ref}
			className={cn(
				"border-border text-md inline-flex h-9 w-full flex-1 rounded-md border bg-transparent px-3 py-2",
				props.isDisabled ? "cursor-not-allowed opacity-50" : "",
				props.className,
			)}
		>
			{state.segments.map((segment, i) => (
				<DateSegment key={i} segment={segment} state={state} />
			))}
		</div>
	);
}

export { TimeField };
