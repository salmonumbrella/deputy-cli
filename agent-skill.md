---
name: deputy-cli
description: Use when interacting with Deputy workforce management API via CLI
---

# Deputy CLI Agent Guide

## Quick Reference

### Authentication
```bash
deputy auth login           # Browser-based OAuth login
deputy auth test            # Verify credentials work
deputy auth status          # Show current auth state
```

### Common Patterns

**List resources:**
```bash
deputy employees list       # or: deputy list employees
deputy locations list       # or: deputy list locations
deputy timesheets list --employee 123
```

**Get by ID:**
```bash
deputy employees get 123    # or: deputy get employee 123
deputy locations get 1
```

**Query any resource:**
```bash
deputy resource query Employee --filter "Active=1"
deputy resource query Timesheet --filter "EmployeeId=123" --filter "Date>=2024-01-01"
```

### Aliases
| Full Command | Shortcuts |
|--------------|-----------|
| employees | employee, emp, e |
| locations | location, loc |
| timesheets | timesheet, ts, t |
| rosters | roster, shifts, shift, r |
| departments | department, dept, d, areas, area |

### Output Formats
```bash
-o json                     # JSON output for parsing
-o json -q '.[] | .Id'      # With jq filter
--raw                       # JSON Lines (one object per line)
--version                   # Show version info
```

### Resource Schema Discovery
```bash
deputy resource list                    # List all resource types
deputy resource info EmployeeAgreement  # Show fields and associations
```

### Pay Management
```bash
deputy pay awards list                          # List available awards
deputy pay agreements list --employee 123       # Employee's pay agreements
deputy pay agreements get 194                   # Get agreement details
deputy pay agreements update 194 --base-rate 25.50  # Update base rate
```

### Common Field Names (for queries)
| Resource | Key Fields |
|----------|------------|
| Employee | Id, Active, DisplayName, FirstName, LastName |
| EmployeeAgreement | Id, EmployeeId, Active, BaseRate, Config |
| Timesheet | Id, EmployeeId, Date, StartTime, EndTime |
| Company | Id, CompanyName, Active |
| OperationalUnit | Id, OperationalUnitName, Company |

### Error Recovery
| Error | Solution |
|-------|----------|
| 401 unauthorized | Run `deputy auth login` |
| 400 Invalid search field | Check field name with `deputy resource info <Resource>` |
| 404 not found | Verify resource exists with `deputy resource list` |
| 417 Unable to cast | Config must be JSON object, not array |
