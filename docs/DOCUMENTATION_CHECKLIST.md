# Documentation Update Checklist

This checklist should be reviewed whenever making code changes to ensure documentation stays synchronized with the implementation.

## When to Update Documentation

### Always Update When:
- [ ] **Adding/removing error types** → Update `docs/DESIGN_PATTERNS.md`
- [ ] **Changing interfaces** → Update `docs/DESIGN_PATTERNS.md` and `docs/IMPLEMENTATION_NOTES.md`
- [ ] **Adding/removing modules** → Update `docs/IMPLEMENTATION_NOTES.md` code organization section
- [ ] **Changing CLI flags/options** → Update `README.md` CLI Options section and `docs/IMPLEMENTATION_NOTES.md`
- [ ] **Modifying error/output messages** → Update `docs/OUTPUT_SPECIFICATION.md`
- [ ] **Changing exit codes** → Update `docs/DESIGN_PATTERNS.md` exit code table
- [ ] **Adding/removing dependencies** → Update `CLAUDE.md` and `docs/IMPLEMENTATION_NOTES.md`

## Documentation Files to Check

### Core Documentation
- [ ] `README.md` - User-facing documentation, examples, CLI options, troubleshooting
- [ ] `CLAUDE.md` - Development guidelines, patterns, commands
- [ ] `DESIGN.md` - High-level architecture and module overview (if present)

### Developer Documentation
- [ ] `docs/DESIGN_PATTERNS.md` - Error handling patterns, orchestration, testing strategies
- [ ] `docs/OUTPUT_SPECIFICATION.md` - Output formats, message specifications, user experience
- [ ] `docs/IMPLEMENTATION_NOTES.md` - Technical constraints, interfaces, platform considerations
- [ ] `docs/ERROR_HANDLING_GUIDE.md` - Best practices guide for Go error handling
- [ ] `docs/DOCUMENTATION_CHECKLIST.md` - This checklist (update when documentation structure changes)

## Update Process

1. **Before Making Changes**:
   - Review relevant documentation to understand current behavior
   - Note which sections will need updates

2. **While Making Changes**:
   - Keep a list of documentation updates needed
   - Update interface definitions immediately
   - Update error types/messages as you change them

3. **After Making Changes**:
   - Review this checklist
   - Update all affected documentation
   - Ensure examples still work
   - Verify interface signatures match implementation

4. **Before Committing**:
   - Run through the checklist one more time
   - Ensure all documentation is in sync
   - Test any command examples in docs

## Common Patterns to Document

### Error Handling Changes
- Error type definitions → `docs/DESIGN_PATTERNS.md`
- Error creation patterns (New* vs Wrap*) → `docs/DESIGN_PATTERNS.md`
- Exit code mappings → `docs/DESIGN_PATTERNS.md`
- Help display behavior → `docs/DESIGN_PATTERNS.md`
- Error message formats → `docs/OUTPUT_SPECIFICATION.md`

### Interface Changes
- Method signatures → `docs/IMPLEMENTATION_NOTES.md`
- Parameter types → `docs/IMPLEMENTATION_NOTES.md`
- Return values → `docs/IMPLEMENTATION_NOTES.md`
- New methods added → `docs/IMPLEMENTATION_NOTES.md`
- Methods removed → `docs/IMPLEMENTATION_NOTES.md`

### CLI Changes
- New flags/options → `README.md` + `docs/IMPLEMENTATION_NOTES.md`
- Changed flag behavior → `README.md` + `docs/IMPLEMENTATION_NOTES.md`
- New commands/subcommands → `README.md` + `docs/IMPLEMENTATION_NOTES.md`
- Changed validation rules → `docs/IMPLEMENTATION_NOTES.md`

### Output Changes
- Success/failure message formats → `docs/OUTPUT_SPECIFICATION.md`
- Progress display changes → `docs/OUTPUT_SPECIFICATION.md`
- Error message wording → `docs/OUTPUT_SPECIFICATION.md`
- ANSI/formatting changes → `docs/OUTPUT_SPECIFICATION.md`

## Documentation Quality Checks

- [ ] Code examples compile and run
- [ ] Interface definitions match actual code
- [ ] Error messages in examples match actual output
- [ ] CLI examples produce expected results
- [ ] No references to removed features
- [ ] All new features are documented

## Current Documentation Structure

After consolidation, the documentation is organized as:

- **User Documentation**: `README.md` (installation, usage, examples, troubleshooting)
- **Developer Onboarding**: `CLAUDE.md` (development setup, commands, patterns)
- **Design Authority**: `docs/DESIGN_PATTERNS.md` (error handling, orchestration, testing)
- **Output Authority**: `docs/OUTPUT_SPECIFICATION.md` (user experience, message formats)
- **Technical Reference**: `docs/IMPLEMENTATION_NOTES.md` (interfaces, constraints, platform)
- **Best Practices**: `docs/ERROR_HANDLING_GUIDE.md` (Go error handling guidance)

## Reminder

**The documentation is the authoritative specification** - it should be updated BEFORE or WITH code changes, not after. When in doubt, update the documentation!