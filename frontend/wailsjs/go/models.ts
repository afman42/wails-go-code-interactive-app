export namespace main {
	
	export class Data {
	    txt: string;
	    out: string;
	    errout: string;
	    lang: string;
	    type: string;
	    execMode: string;
	    bundledRuntime: string;
	    customExecutable: string;
	    customWorkingDir: string;
	    preferBundled: boolean;
	
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
	        this.execMode = source["execMode"];
	        this.bundledRuntime = source["bundledRuntime"];
	        this.customExecutable = source["customExecutable"];
	        this.customWorkingDir = source["customWorkingDir"];
	        this.preferBundled = source["preferBundled"];
	    }
	}
	export class LanguageAvailability {
	    system: string[];
	    bundled: string[];
	
	    static createFrom(source: any = {}) {
	        return new LanguageAvailability(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.system = source["system"];
	        this.bundled = source["bundled"];
	    }
	}

}

