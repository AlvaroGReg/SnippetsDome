export namespace domain {
	
	export class Collection {
	    id: string;
	    name: string;
	    createdAt: string;
	    encrypted: boolean;
	    locked: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Collection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.createdAt = source["createdAt"];
	        this.encrypted = source["encrypted"];
	        this.locked = source["locked"];
	    }
	}
	export class CreateSnippetInput {
	    title: string;
	    language: string;
	    code: string;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateSnippetInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.language = source["language"];
	        this.code = source["code"];
	        this.tags = source["tags"];
	    }
	}
	export class Snippet {
	    id: string;
	    title: string;
	    language: string;
	    code: string;
	    tags: string[];
	    createdAt: string;
	    favorite: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Snippet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.language = source["language"];
	        this.code = source["code"];
	        this.tags = source["tags"];
	        this.createdAt = source["createdAt"];
	        this.favorite = source["favorite"];
	    }
	}

}

export namespace main {
	
	export class JSONImportResult {
	    needsPassword: boolean;
	    collection?: domain.Collection;
	
	    static createFrom(source: any = {}) {
	        return new JSONImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.needsPassword = source["needsPassword"];
	        this.collection = this.convertValues(source["collection"], domain.Collection);
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

