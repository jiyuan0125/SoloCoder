import { v4 as uuidv4 } from 'uuid';
import { db, POVERTY_LEVELS, POVERTY_CAUSES, POVERTY_LINE } from '../database';
import type { Household, PovertyLevel, PovertyCause, HouseholdStatus } from '../types';

function parseHousehold(row: any): Household {
  return {
    id: row.id,
    householdId: row.household_id,
    county: row.county,
    township: row.township,
    village: row.village,
    headOfHousehold: row.head_of_household,
    contactPhone: row.contact_phone,
    familySize: row.family_size,
    familyMembers: JSON.parse(row.family_members),
    povertyCauses: JSON.parse(row.poverty_causes),
    mainPovertyCause: row.main_poverty_cause,
    povertyLevel: row.poverty_level,
    assets: JSON.parse(row.assets),
    laborForce: JSON.parse(row.labor_force),
    yearlyIncomePerCapita: row.yearly_income_per_capita,
    twoWorriesThreeGuarantees: JSON.parse(row.two_worries_three_guarantees),
    status: row.status,
    outOfPovertyDate: row.out_of_poverty_date,
    monitoringEndDate: row.monitoring_end_date,
    helperName: row.helper_name,
    helperUnit: row.helper_unit,
    helperPhone: row.helper_phone,
    archiveYear: row.archive_year,
    createdAt: row.created_at,
    updatedAt: row.updated_at
  };
}

function addHouseholdHistory(
  householdId: string,
  eventType: 'archive_update' | 'status_change' | 'out_of_poverty' | 'return_to_poverty',
  description: string,
  operator: string,
  previousStatus?: HouseholdStatus,
  newStatus?: HouseholdStatus
): void {
  const stmt = db.prepare(`
    INSERT INTO household_histories (id, household_id, event_type, previous_status, new_status, description, operator, operation_date)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `);
  stmt.run(
    uuidv4(),
    householdId,
    eventType,
    previousStatus || null,
    newStatus || null,
    description,
    operator,
    new Date().toISOString()
  );
}

function checkPovertyEligibility(
  yearlyIncome: number,
  twoWorries: any
): boolean {
  const incomeOk = yearlyIncome > POVERTY_LINE;
  const twoWorriesOk = 
    !twoWorries.worryAboutFood &&
    !twoWorries.worryAboutClothing &&
    twoWorries.compulsoryEducation &&
    twoWorries.basicMedicalCare &&
    twoWorries.housingSafety;
  return incomeOk && twoWorriesOk;
}

export function createHousehold(data: any, operator: string): Household {
  if (!data.povertyCauses || data.povertyCauses.length === 0) {
    throw new Error('POVERTY_CAUSES_REQUIRED');
  }
  
  const invalidCauses = data.povertyCauses.filter((c: string) => !POVERTY_CAUSES.includes(c as PovertyCause));
  if (invalidCauses.length > 0) {
    throw new Error('INVALID_POVERTY_CAUSE');
  }

  if (!POVERTY_LEVELS.includes(data.povertyLevel)) {
    throw new Error('INVALID_POVERTY_LEVEL');
  }

  const now = new Date().toISOString();
  const id = uuidv4();
  const householdId = data.householdId || `HH${Date.now()}`;
  const mainPovertyCause = data.povertyCauses[0];

  const stmt = db.prepare(`
    INSERT INTO households (
      id, household_id, county, township, village, head_of_household, contact_phone,
      family_size, family_members, poverty_causes, main_poverty_cause,
      poverty_level, assets, labor_force, yearly_income_per_capita, two_worries_three_guarantees,
      status, helper_name, helper_unit, helper_phone, archive_year, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);

  stmt.run(
    id,
    householdId,
    data.county,
    data.township,
    data.village,
    data.headOfHousehold,
    data.contactPhone || null,
    data.familySize,
    JSON.stringify(data.familyMembers || []),
    JSON.stringify(data.povertyCauses),
    mainPovertyCause,
    data.povertyLevel,
    JSON.stringify(data.assets || {}),
    JSON.stringify(data.laborForce || {}),
    data.yearlyIncomePerCapita || 0,
    JSON.stringify(data.twoWorriesThreeGuarantees || {
      worryAboutFood: true,
      worryAboutClothing: true,
      compulsoryEducation: false,
      basicMedicalCare: false,
      housingSafety: false
    }),
    'poverty',
    data.helperName,
    data.helperUnit,
    data.helperPhone || null,
    new Date().getFullYear(),
    now,
    now
  );

  addHouseholdHistory(
    householdId,
    'archive_update',
    '创建贫困户档案',
    operator
  );

  return getHouseholdById(householdId)!;
}

export function getHouseholdById(householdId: string): Household | null {
  const row = db.prepare('SELECT * FROM households WHERE household_id = ?').get(householdId);
  return row ? parseHousehold(row) : null;
}

export function updateArchive(householdId: string, data: any, operator: string): Household {
  const existing = getHouseholdById(householdId);
  if (!existing) {
    throw new Error('HOUSEHOLD_NOT_FOUND');
  }

  if (data.povertyCauses) {
    if (data.povertyCauses.length === 0) {
      throw new Error('POVERTY_CAUSES_REQUIRED');
    }
    const invalidCauses = data.povertyCauses.filter((c: string) => !POVERTY_CAUSES.includes(c as PovertyCause));
    if (invalidCauses.length > 0) {
      throw new Error('INVALID_POVERTY_CAUSE');
    }
  }

  if (data.povertyLevel && !POVERTY_LEVELS.includes(data.povertyLevel)) {
    throw new Error('INVALID_POVERTY_LEVEL');
  }

  const povertyCauses = data.povertyCauses || existing.povertyCauses;
  const mainPovertyCause = povertyCauses[0];

  const now = new Date().toISOString();

  const stmt = db.prepare(`
    UPDATE households SET
      county = COALESCE(?, county),
      township = COALESCE(?, township),
      village = COALESCE(?, village),
      head_of_household = COALESCE(?, head_of_household),
      contact_phone = COALESCE(?, contact_phone),
      family_size = COALESCE(?, family_size),
      family_members = COALESCE(?, family_members),
      poverty_causes = COALESCE(?, poverty_causes),
      main_poverty_cause = ?,
      poverty_level = COALESCE(?, poverty_level),
      assets = COALESCE(?, assets),
      labor_force = COALESCE(?, labor_force),
      yearly_income_per_capita = COALESCE(?, yearly_income_per_capita),
      two_worries_three_guarantees = COALESCE(?, two_worries_three_guarantees),
      helper_name = COALESCE(?, helper_name),
      helper_unit = COALESCE(?, helper_unit),
      helper_phone = COALESCE(?, helper_phone),
      archive_year = ?,
      updated_at = ?
    WHERE household_id = ?
  `);

  stmt.run(
    data.county,
    data.township,
    data.village,
    data.headOfHousehold,
    data.contactPhone,
    data.familySize,
    data.familyMembers ? JSON.stringify(data.familyMembers) : null,
    data.povertyCauses ? JSON.stringify(data.povertyCauses) : null,
    mainPovertyCause,
    data.povertyLevel,
    data.assets ? JSON.stringify(data.assets) : null,
    data.laborForce ? JSON.stringify(data.laborForce) : null,
    data.yearlyIncomePerCapita,
    data.twoWorriesThreeGuarantees ? JSON.stringify(data.twoWorriesThreeGuarantees) : null,
    data.helperName,
    data.helperUnit,
    data.helperPhone,
    new Date().getFullYear(),
    now,
    householdId
  );

  addHouseholdHistory(
    householdId,
    'archive_update',
    `年度档案更新`,
    operator
  );

  const updated = getHouseholdById(householdId)!;
  
  const shouldBeOut = checkPovertyEligibility(
    updated.yearlyIncomePerCapita,
    updated.twoWorriesThreeGuarantees
  );

  if (shouldBeOut && updated.status === 'poverty') {
    markOutOfPoverty(householdId, operator);
  }

  return getHouseholdById(householdId)!;
}

export function markOutOfPoverty(householdId: string, operator: string): Household {
  const existing = getHouseholdById(householdId);
  if (!existing) {
    throw new Error('HOUSEHOLD_NOT_FOUND');
  }

  const now = new Date();
  const outDate = now.toISOString();
  const monitoringEnd = new Date(now.getFullYear() + 2, now.getMonth(), now.getDate()).toISOString();

  const previousStatus = existing.status;

  db.prepare(`
    UPDATE households SET
      status = 'out_of_poverty',
      out_of_poverty_date = ?,
      monitoring_end_date = ?,
      updated_at = ?
    WHERE household_id = ?
  `).run(outDate, monitoringEnd, outDate, householdId);

  addHouseholdHistory(
    householdId,
    'out_of_poverty',
    '贫困户脱贫，开始两年跟踪期',
    operator,
    previousStatus,
    'out_of_poverty'
  );

  return getHouseholdById(householdId)!;
}

export function markReturnMonitoring(householdId: string, operator: string): Household {
  const existing = getHouseholdById(householdId);
  if (!existing) {
    throw new Error('HOUSEHOLD_NOT_FOUND');
  }

  if (existing.status !== 'out_of_poverty') {
    throw new Error('NOT_IN_MONITORING_PERIOD');
  }

  const now = new Date().toISOString();
  const previousStatus = existing.status;

  db.prepare(`
    UPDATE households SET
      status = 'return_monitoring',
      updated_at = ?
    WHERE household_id = ?
  `).run(now, householdId);

  addHouseholdHistory(
    householdId,
    'return_to_poverty',
    '返贫监测',
    operator,
    previousStatus,
    'return_monitoring'
  );

  return getHouseholdById(householdId)!;
}

export function getAllHouseholds(): Household[] {
  const rows = db.prepare('SELECT * FROM households').all();
  return rows.map(parseHousehold);
}

export function getHouseholdHistories(householdId: string): any[] {
  return db.prepare(`
    SELECT * FROM household_histories 
    WHERE household_id = ? 
    ORDER BY operation_date DESC
  `).all(householdId);
}
