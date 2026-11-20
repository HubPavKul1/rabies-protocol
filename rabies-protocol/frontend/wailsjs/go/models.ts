export namespace domain {
	
	export class ClinicConfig {
	    name: string;
	    address: string;
	    phone: string;
	    email: string;
	
	    static createFrom(source: any = {}) {
	        return new ClinicConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.address = source["address"];
	        this.phone = source["phone"];
	        this.email = source["email"];
	    }
	}
	export class Veterinarian {
	    id: string;
	    full_name: string;
	    position: string;
	
	    static createFrom(source: any = {}) {
	        return new Veterinarian(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.full_name = source["full_name"];
	        this.position = source["position"];
	    }
	}

}

