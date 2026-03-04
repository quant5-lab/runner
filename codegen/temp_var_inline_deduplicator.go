package codegen

type TempVarInlineDeduplicator struct {
	tempVarMgr *TempVariableManager
}

func NewTempVarInlineDeduplicator(mgr *TempVariableManager) *TempVarInlineDeduplicator {
	return &TempVarInlineDeduplicator{
		tempVarMgr: mgr,
	}
}

func (d *TempVarInlineDeduplicator) ShouldEmitCalculation(callInfo CallInfo) bool {
	varName := d.tempVarMgr.GetVarNameForCall(callInfo.Call)
	if varName == "" {
		return true
	}
	return !d.tempVarMgr.WasAlreadyEmitted(varName)
}
