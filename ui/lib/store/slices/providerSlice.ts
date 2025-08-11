import { ProviderResponse } from "@/lib/types/config";
import { createSlice, PayloadAction } from "@reduxjs/toolkit";

interface ProviderState {
	selectedProvider: ProviderResponse | null;
	isConfigureDialogOpen: boolean;
	providers: ProviderResponse[];
}

const initialState: ProviderState = {
	selectedProvider: null,
	isConfigureDialogOpen: false,
	providers: [],
};

const providerSlice = createSlice({
	name: "provider",
	initialState,
	reducers: {
		setSelectedProvider: (state, action: PayloadAction<ProviderResponse | null>) => {
			state.selectedProvider = action.payload;
		},
		setIsConfigureDialogOpen: (state, action: PayloadAction<boolean>) => {
			state.isConfigureDialogOpen = action.payload;
		},
		setProviders: (state, action: PayloadAction<ProviderResponse[]>) => {
			state.providers = action.payload;
		},
		openConfigureDialog: (state, action: PayloadAction<ProviderResponse | null>) => {
			state.selectedProvider = action.payload;
			state.isConfigureDialogOpen = true;
		},
		closeConfigureDialog: (state) => {
			state.selectedProvider = null;
			state.isConfigureDialogOpen = false;
		},
	},
});

export const { setSelectedProvider, setIsConfigureDialogOpen, setProviders, openConfigureDialog, closeConfigureDialog } =
	providerSlice.actions;

export default providerSlice.reducer;
