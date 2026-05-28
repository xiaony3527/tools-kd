export namespace main {
	
	export class AddressResult {
	    province: string;
	    city: string;
	
	    static createFrom(source: any = {}) {
	        return new AddressResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.province = source["province"];
	        this.city = source["city"];
	    }
	}
	export class CarrierQuote {
	    price: number;
	    billableWeight: number;
	    available: boolean;
	    recommended: boolean;
	    note: string;
	
	    static createFrom(source: any = {}) {
	        return new CarrierQuote(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.price = source["price"];
	        this.billableWeight = source["billableWeight"];
	        this.available = source["available"];
	        this.recommended = source["recommended"];
	        this.note = source["note"];
	    }
	}
	export class DestinationState {
	    province: string;
	    city: string;
	    cities: string[];
	
	    static createFrom(source: any = {}) {
	        return new DestinationState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.province = source["province"];
	        this.city = source["city"];
	        this.cities = source["cities"];
	    }
	}
	export class PackageInput {
	    length: number;
	    width: number;
	    height: number;
	    volume: number;
	    actualWeight: number;
	    quantity: number;
	
	    static createFrom(source: any = {}) {
	        return new PackageInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.length = source["length"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.volume = source["volume"];
	        this.actualWeight = source["actualWeight"];
	        this.quantity = source["quantity"];
	    }
	}
	export class PackageRow {
	    id: number;
	    length: number;
	    width: number;
	    height: number;
	    quantity: number;
	    mode: number;
	    modeLabel: string;
	    actualWeight: number;
	    volume: number;
	    stoBillable: number;
	    bsBillable: number;
	
	    static createFrom(source: any = {}) {
	        return new PackageRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.length = source["length"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.quantity = source["quantity"];
	        this.mode = source["mode"];
	        this.modeLabel = source["modeLabel"];
	        this.actualWeight = source["actualWeight"];
	        this.volume = source["volume"];
	        this.stoBillable = source["stoBillable"];
	        this.bsBillable = source["bsBillable"];
	    }
	}
	export class SummaryState {
	    totalCount: number;
	    displayWeight: number;
	
	    static createFrom(source: any = {}) {
	        return new SummaryState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalCount = source["totalCount"];
	        this.displayWeight = source["displayWeight"];
	    }
	}
	export class QuoteState {
	    destination: DestinationState;
	    packages: PackageRow[];
	    summary: SummaryState;
	    sto: CarrierQuote;
	    bs: CarrierQuote;
	
	    static createFrom(source: any = {}) {
	        return new QuoteState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.destination = this.convertValues(source["destination"], DestinationState);
	        this.packages = this.convertValues(source["packages"], PackageRow);
	        this.summary = this.convertValues(source["summary"], SummaryState);
	        this.sto = this.convertValues(source["sto"], CarrierQuote);
	        this.bs = this.convertValues(source["bs"], CarrierQuote);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

