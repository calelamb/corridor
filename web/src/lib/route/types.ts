export type RiskTier = 'Low' | 'Moderate' | 'High' | 'Severe';

export interface HazardZone {
	id: string;
	segmentId?: string;
	highwayRef?: string;
	highway_ref?: string;
	milepost: string | number;
	primaryRiskReason?: string;
	primary_risk_reason?: string;
	riskScore?: number;
	risk_score?: number;
	riskTier?: RiskTier | string;
}

export interface RouteRiskReport {
	overallRiskTier: RiskTier | string;
	averageRiskScore: number;
	peakRiskScore: number;
	diurnalMultiplier: number;
	isPeakWildlifeHour: boolean;
	hazardZones: HazardZone[];
	advisoryNotice?: string;
	totalDistanceKm: number;
}
