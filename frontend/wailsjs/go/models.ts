export namespace gui {
	
	export class ConversionResult {
	    success: boolean;
	    outputPath?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConversionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.outputPath = source["outputPath"];
	        this.error = source["error"];
	    }
	}

}

