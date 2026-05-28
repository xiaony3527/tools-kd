// Re-export all Wails-bound App methods so consumers import from './api' instead of deep paths.
import {
  GetInitialState,
  AddPackage,
  DeletePackage,
  ClearPackages,
  SetDestination,
  AnalyzeAddress,
  CalculateQuote,
  GetProvinces,
} from './wailsjs/go/main/App';

export {
  GetInitialState,
  AddPackage,
  DeletePackage,
  ClearPackages,
  SetDestination,
  AnalyzeAddress,
  CalculateQuote,
  GetProvinces,
};
