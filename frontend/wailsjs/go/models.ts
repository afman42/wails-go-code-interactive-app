export namespace main {
	
	export class Data {
	    txt: string;
	    out: string;
	    errout: string;
	    lang: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new Data(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.txt = source["txt"];
	        this.out = source["out"];
	        this.errout = source["errout"];
	        this.lang = source["lang"];
	        this.type = source["type"];
	    }
	}

}

