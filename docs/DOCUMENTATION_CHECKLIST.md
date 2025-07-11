# Documentation Update Checklist

This checklist should be reviewed whenever making code changes to ensure documentation stays synchronized with the implementation.

## When to Update Documentation

### Always Update When:
- [ ] **Adding/removing error types** → Update `docs/features/error-handling.md`
- [ ] **Changing interfaces** → Update `DESIGN.md` and relevant feature docs
- [ ] **Adding/removing modules** → Update `DESIGN.md` directory structure
- [ ] **Changing CLI flags/options** → Update `README.md` CLI Options section
- [ ] **Modifying error messages** → Update error examples in documentation
- [ ] **Changing exit codes** → Update exit code tables
- [ ] **Adding/removing dependencies** → Update `CLAUDE.md` dependencies section

## Documentation Files to Check

### Core Documentation
- [ ] `DESIGN.md` - Architecture, modules, interfaces, execution flow
- [ ] `README.md` - User-facing documentation, examples, CLI options
- [ ] `CLAUDE.md` - Development guidelines, patterns, commands

### Feature Documentation
- [ ] `docs/features/error-handling.md` - Error types, exit codes, help behavior
- [ ] `docs/features/git-operations.md` - Git worktree management
- [ ] `docs/features/command-execution.md` - Command executor behavior
- [ ] `docs/features/metric-parsing.md` - Parser rules and examples
- [ ] `docs/features/configuration.md` - Config structure and validation

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
- Error type definitions
- Error creation patterns (New* vs Wrap*)
- Exit code mappings
- Help display behavior
- Error message formats

### Interface Changes
- Method signatures
- Parameter types
- Return values
- New methods added
- Methods removed

### CLI Changes
- New flags/options
- Changed flag behavior
- New commands/subcommands
- Changed validation rules

## Documentation Quality Checks

- [ ] Code examples compile and run
- [ ] Interface definitions match actual code
- [ ] Error messages in examples match actual output
- [ ] CLI examples produce expected results
- [ ] No references to removed features
- [ ] All new features are documented

## Reminder

**The documentation is the authoritative specification** - it should be updated BEFORE or WITH code changes, not after. When in doubt, update the documentation!