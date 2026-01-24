export interface M3Config {
  tenantId: string;
  instanceId: string;
  environment: string;
}

/**
 * Builds an M3 CloudSuite bookmark URL for a production order
 * @param config - M3 configuration (tenant ID, instance ID)
 * @param orderType - 'MO' for Manufacturing Order or 'MOP' for Planned MO
 * @param orderNumber - The order number (MFNO for MO, PLPN for MOP)
 * @param company - Company number (CONO)
 * @param facility - Facility code (FACI) - optional for MOPs, required for MOs
 * @param productNumber - Product number (PRNO) - optional, used for MOs
 * @returns Complete M3 bookmark URL
 */
export function buildM3BookmarkURL(
  config: M3Config,
  orderType: 'MO' | 'MOP',
  orderNumber: string,
  company: string,
  facility?: string,
  productNumber?: string
): string {
  const baseUrl = `https://mingle-portal.inforcloudsuite.com/${config.tenantId}/${config.instanceId}`;

  let favoriteContext: string;

  if (orderType === 'MO') {
    // PMS100 - Manufacturing Order
    // M3 product numbers are 15 characters, right-padded with + signs
    const paddedProduct = productNumber ? productNumber.padEnd(15, '+') : null;

    // Format: bookmark?PMS100&MWOHED&VHCONO%2C{company}%2CVHFACI%2C{facility}%2CVHPRNO%2C{product}%2CVHMFNO%2C{moNumber}&5&E&PMS100/E&PMS100 Manufacturing Order. Open
    const fields = [
      `VHCONO%2C${company}`,
      facility ? `%2CVHFACI%2C${facility}` : null,
      paddedProduct ? `%2CVHPRNO%2C${paddedProduct}` : null,
      `%2CVHMFNO%2C${orderNumber}`,
    ]
      .filter(Boolean)
      .join('');

    favoriteContext = `bookmark?PMS100&MWOHED&${fields}&5&E&PMS100/E&PMS100 Manufacturing Order. Open`;
  } else {
    // PMS170 - Planned Manufacturing Order
    // Format: bookmark?PMS170&MMOPLP&ROCONO%2C{company}%2CROPLPN%2C{mopNumber}%2CROPLPS%2C0&5&E&PMS170/E&PMS170 Planned MO. Open
    favoriteContext = `bookmark?PMS170&MMOPLP&ROCONO%2C${company}%2CROPLPN%2C${orderNumber}%2CROPLPS%2C0&5&E&PMS170/E&PMS170 Planned MO. Open`;
  }

  const encodedContext = encodeURIComponent(favoriteContext);
  // Pre-encode the colon and slashes in lid://infor.m3.m3
  const logicalId = 'lid%3A%2F%2Finfor.m3.m3';

  return `${baseUrl}?favoriteContext=${encodedContext}&LogicalId=${logicalId}`;
}
