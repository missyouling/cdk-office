# Scripts

This directory contains various scripts for the CDK project.

## Available Scripts

- [license.sh](license.sh) - License header check and update script
- [swagger.sh](swagger.sh) - Swagger documentation generation script
- [tidy.sh](tidy.sh) - Go mod tidy script
- [deploy-vps.sh](deploy-vps.sh) - VPS optimized deployment script for 2C4G environments
- [stop-vps.sh](stop-vps.sh) - Script to stop services deployed with deploy-vps.sh

## VPS Deployment

For 2C4G VPS environments, use the optimized deployment scripts:

```bash
# Deploy with memory optimization
./scripts/deploy-vps.sh

# Stop the services
./scripts/stop-vps.sh
```

The VPS deployment script will:
1. Build an optimized binary with reduced memory footprint
2. Apply memory-optimized configuration from `config.vps.yaml`
3. Set Go memory optimization environment variables
4. Start both API and Worker services in the background
5. Save process IDs for easy management

## Environment Variables

The VPS deployment sets these Go memory optimization variables:
- `GOGC=20` - More aggressive garbage collection
- `GOMEMLIMIT=3GiB` - Set memory limit to prevent OOM