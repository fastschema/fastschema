(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[598],{2071:function(e,l,n){Promise.resolve().then(n.bind(n,12428))},26723:function(e){e.exports=function(){var e=[],l=[],n={},s={},a={};function r(e){return"string"==typeof e?RegExp("^"+e+"$","i"):e}function t(e,l){return e===l?l:e===e.toLowerCase()?l.toLowerCase():e===e.toUpperCase()?l.toUpperCase():e[0]===e[0].toUpperCase()?l.charAt(0).toUpperCase()+l.substr(1).toLowerCase():l.toLowerCase()}function i(e,l,s){if(!e.length||n.hasOwnProperty(e))return l;for(var a=s.length;a--;){var r=s[a];if(r[0].test(l))return function(e,l){return e.replace(l[0],function(n,s){var a,r,i=(a=l[1],r=arguments,a.replace(/\$(\d{1,2})/g,function(e,l){return r[l]||""}));return""===n?t(e[s-1],i):t(n,i)})}(l,r)}return l}function o(e,l,n){return function(s){var a=s.toLowerCase();return l.hasOwnProperty(a)?t(s,a):e.hasOwnProperty(a)?t(s,e[a]):i(a,s,n)}}function d(e,l,n,s){return function(s){var a=s.toLowerCase();return!!l.hasOwnProperty(a)||!e.hasOwnProperty(a)&&i(a,a,n)===a}}function c(e,l,n){var s=1===l?c.singular(e):c.plural(e);return(n?l+" ":"")+s}return c.plural=o(a,s,e),c.isPlural=d(a,s,e),c.singular=o(s,a,l),c.isSingular=d(s,a,l),c.addPluralRule=function(l,n){e.push([r(l),n])},c.addSingularRule=function(e,n){l.push([r(e),n])},c.addUncountableRule=function(e){if("string"==typeof e){n[e.toLowerCase()]=!0;return}c.addPluralRule(e,"$0"),c.addSingularRule(e,"$0")},c.addIrregularRule=function(e,l){l=l.toLowerCase(),a[e=e.toLowerCase()]=l,s[l]=e},[["I","we"],["me","us"],["he","they"],["she","they"],["them","them"],["myself","ourselves"],["yourself","yourselves"],["itself","themselves"],["herself","themselves"],["himself","themselves"],["themself","themselves"],["is","are"],["was","were"],["has","have"],["this","these"],["that","those"],["echo","echoes"],["dingo","dingoes"],["volcano","volcanoes"],["tornado","tornadoes"],["torpedo","torpedoes"],["genus","genera"],["viscus","viscera"],["stigma","stigmata"],["stoma","stomata"],["dogma","dogmata"],["lemma","lemmata"],["schema","schemata"],["anathema","anathemata"],["ox","oxen"],["axe","axes"],["die","dice"],["yes","yeses"],["foot","feet"],["eave","eaves"],["goose","geese"],["tooth","teeth"],["quiz","quizzes"],["human","humans"],["proof","proofs"],["carve","carves"],["valve","valves"],["looey","looies"],["thief","thieves"],["groove","grooves"],["pickaxe","pickaxes"],["passerby","passersby"]].forEach(function(e){return c.addIrregularRule(e[0],e[1])}),[[/s?$/i,"s"],[/[^\u0000-\u007F]$/i,"$0"],[/([^aeiou]ese)$/i,"$1"],[/(ax|test)is$/i,"$1es"],[/(alias|[^aou]us|t[lm]as|gas|ris)$/i,"$1es"],[/(e[mn]u)s?$/i,"$1s"],[/([^l]ias|[aeiou]las|[ejzr]as|[iu]am)$/i,"$1"],[/(alumn|syllab|vir|radi|nucle|fung|cact|stimul|termin|bacill|foc|uter|loc|strat)(?:us|i)$/i,"$1i"],[/(alumn|alg|vertebr)(?:a|ae)$/i,"$1ae"],[/(seraph|cherub)(?:im)?$/i,"$1im"],[/(her|at|gr)o$/i,"$1oes"],[/(agend|addend|millenni|dat|extrem|bacteri|desiderat|strat|candelabr|errat|ov|symposi|curricul|automat|quor)(?:a|um)$/i,"$1a"],[/(apheli|hyperbat|periheli|asyndet|noumen|phenomen|criteri|organ|prolegomen|hedr|automat)(?:a|on)$/i,"$1a"],[/sis$/i,"ses"],[/(?:(kni|wi|li)fe|(ar|l|ea|eo|oa|hoo)f)$/i,"$1$2ves"],[/([^aeiouy]|qu)y$/i,"$1ies"],[/([^ch][ieo][ln])ey$/i,"$1ies"],[/(x|ch|ss|sh|zz)$/i,"$1es"],[/(matr|cod|mur|sil|vert|ind|append)(?:ix|ex)$/i,"$1ices"],[/\b((?:tit)?m|l)(?:ice|ouse)$/i,"$1ice"],[/(pe)(?:rson|ople)$/i,"$1ople"],[/(child)(?:ren)?$/i,"$1ren"],[/eaux$/i,"$0"],[/m[ae]n$/i,"men"],["thou","you"]].forEach(function(e){return c.addPluralRule(e[0],e[1])}),[[/s$/i,""],[/(ss)$/i,"$1"],[/(wi|kni|(?:after|half|high|low|mid|non|night|[^\w]|^)li)ves$/i,"$1fe"],[/(ar|(?:wo|[ae])l|[eo][ao])ves$/i,"$1f"],[/ies$/i,"y"],[/\b([pl]|zomb|(?:neck|cross)?t|coll|faer|food|gen|goon|group|lass|talk|goal|cut)ies$/i,"$1ie"],[/\b(mon|smil)ies$/i,"$1ey"],[/\b((?:tit)?m|l)ice$/i,"$1ouse"],[/(seraph|cherub)im$/i,"$1"],[/(x|ch|ss|sh|zz|tto|go|cho|alias|[^aou]us|t[lm]as|gas|(?:her|at|gr)o|[aeiou]ris)(?:es)?$/i,"$1"],[/(analy|diagno|parenthe|progno|synop|the|empha|cri|ne)(?:sis|ses)$/i,"$1sis"],[/(movie|twelve|abuse|e[mn]u)s$/i,"$1"],[/(test)(?:is|es)$/i,"$1is"],[/(alumn|syllab|vir|radi|nucle|fung|cact|stimul|termin|bacill|foc|uter|loc|strat)(?:us|i)$/i,"$1us"],[/(agend|addend|millenni|dat|extrem|bacteri|desiderat|strat|candelabr|errat|ov|symposi|curricul|quor)a$/i,"$1um"],[/(apheli|hyperbat|periheli|asyndet|noumen|phenomen|criteri|organ|prolegomen|hedr|automat)a$/i,"$1on"],[/(alumn|alg|vertebr)ae$/i,"$1a"],[/(cod|mur|sil|vert|ind)ices$/i,"$1ex"],[/(matr|append)ices$/i,"$1ix"],[/(pe)(rson|ople)$/i,"$1rson"],[/(child)ren$/i,"$1"],[/(eau)x?$/i,"$1"],[/men$/i,"man"]].forEach(function(e){return c.addSingularRule(e[0],e[1])}),["adulthood","advice","agenda","aid","aircraft","alcohol","ammo","analytics","anime","athletics","audio","bison","blood","bream","buffalo","butter","carp","cash","chassis","chess","clothing","cod","commerce","cooperation","corps","debris","diabetes","digestion","elk","energy","equipment","excretion","expertise","firmware","flounder","fun","gallows","garbage","graffiti","hardware","headquarters","health","herpes","highjinks","homework","housework","information","jeans","justice","kudos","labour","literature","machinery","mackerel","mail","media","mews","moose","music","mud","manga","news","only","personnel","pike","plankton","pliers","police","pollution","premises","rain","research","rice","salmon","scissors","series","sewage","shambles","shrimp","software","species","staff","swine","tennis","traffic","transportation","trout","tuna","wealth","welfare","whiting","wildebeest","wildlife","you",/pok[eé]mon$/i,/[^aeiou]ese$/i,/deer$/i,/fish$/i,/measles$/i,/o[iu]s$/i,/pox$/i,/sheep$/i].forEach(c.addUncountableRule),c}()},12428:function(e,l,n){"use strict";n.r(l),n.d(l,{default:function(){return D}});var s=n(57437),a=n(76540),r=n(34560),t=n(57654),i=n(80244),o=n(79984),d=n(27453),c=n(96304),u=n(21270),m=n(47907),h=n(26723),x=n.n(h),f=n(2265),j=n(82670),p=n(99773),v=n(50326),g=n(99497),y=n(32302),b=n(22782),w=n(18641),N=n(70094);let $=e=>{let{form:l}=e,{fields:n,append:a,remove:r}=(0,j.Dq)({name:"enums",control:l.control});return(0,s.jsxs)("div",{className:"space-y-4",children:[(0,s.jsx)("h3",{className:"text-md font-medium",children:"Enums"}),(0,s.jsx)(i.Wi,{control:l.control,name:"type",render:e=>{let{field:l}=e;return(0,s.jsx)(i.xJ,{className:"flex-shrink flex-grow flex-1 relative",children:(0,s.jsx)(i.zG,{})})}}),n.map((e,n)=>(0,s.jsxs)("div",{className:"flex flex-wrap items-stretch w-full relative gap-3",children:[(0,s.jsx)(i.Wi,{control:l.control,name:"enums.".concat(n,".value"),render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{className:"flex-shrink flex-grow flex-1 relative",children:[(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{placeholder:"Enum value",...l})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"enums.".concat(n,".label"),render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{className:"flex-shrink flex-grow flex-1 relative",children:[(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{placeholder:"Enum label",...l})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)("div",{className:"flex",children:(0,s.jsx)(t.z,{variant:"destructive",onClick:()=>{window.confirm("Are you sure you want to delete this enum?")&&r(n)},children:"Delete"})})]},e.value||n)),(0,s.jsxs)(t.z,{type:"button",variant:"outline",size:"sm",className:"mt-2",onClick:()=>a({value:"",label:""}),children:[(0,s.jsx)(N.Z,{size:16,className:"mr-2"}),"Add Enum value"]})]})};var C=n(75006),I=n(86468),_=n(15e3),k=n(82628),S=n(95453);let z=e=>{let{form:l,editingField:n}=e,{data:r,isLoading:t,error:o}=(0,a.a)({queryKey:["schemas"],queryFn:c.cQ});return t?(0,s.jsx)(C.g,{}):o?(0,s.jsx)(_.T,{error:o}):(0,s.jsxs)("div",{className:"space-y-4",children:[(0,s.jsx)(S.Z,{className:"my-4"}),(0,s.jsx)(i.Wi,{control:l.control,name:"type",render:e=>{let{field:l}=e;return(0,s.jsx)(i.xJ,{className:"flex-shrink flex-grow flex-1 relative",children:(0,s.jsx)(i.zG,{})})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"relation.type",render:e=>{let{field:l}=e;return(null==n?void 0:n.name)?(0,s.jsx)(b.I,{...l,readOnly:!0}):(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Relation type"}),(0,s.jsxs)(w.Ph,{onValueChange:l.onChange,defaultValue:l.value,name:"relation.type",children:[(0,s.jsx)(i.NI,{children:(0,s.jsx)(w.i4,{children:(0,s.jsx)(w.ki,{placeholder:"Select a relation type"})})}),(0,s.jsxs)(w.Bw,{children:[(0,s.jsx)(w.Ql,{value:"o2o",children:"O2O"}),(0,s.jsx)(w.Ql,{value:"o2m",children:"O2M"}),(0,s.jsx)(w.Ql,{value:"m2m",children:"M2M"})]})]}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"relation.schema",render:e=>{let{field:l}=e;return(null==n?void 0:n.name)?(0,s.jsx)(b.I,{...l,readOnly:!0}):(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Relation schema"}),(0,s.jsxs)(w.Ph,{onValueChange:l.onChange,defaultValue:l.value,name:"relation.schema",children:[(0,s.jsx)(i.NI,{children:(0,s.jsx)(w.i4,{children:(0,s.jsx)(w.ki,{placeholder:"Select a relation schema"})})}),(0,s.jsx)(w.Bw,{children:null==r?void 0:r.map(e=>(0,s.jsx)(w.Ql,{value:e.name,children:e.name},e.name))})]}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"relation.field",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:(0,s.jsx)(k.u,{icon:!0,tip:(0,s.jsxs)(s.Fragment,{children:[" ",(0,s.jsx)("p",{children:"The relation field is the field name of the relation schema that points back to the current editing field."}),(0,s.jsx)("p",{children:"For example:"}),(0,s.jsxs)("ul",{className:"list-decimal list-inside",children:[(0,s.jsxs)("li",{children:["The current editing field is ",(0,s.jsx)("code",{children:"post.post_meta"}),", and the relation schema is ",(0,s.jsx)("code",{children:"post_meta"})," (o2o), then the relation field is ",(0,s.jsx)("code",{children:"post_meta.post"})," (enter ",(0,s.jsx)("code",{children:"post"})," here)"]}),(0,s.jsxs)("li",{children:["The current editing field is ",(0,s.jsx)("code",{children:"post.comments"}),", and the relation schema is ",(0,s.jsx)("code",{children:"comment"})," (o2m), then the relation field is ",(0,s.jsx)("code",{children:"comment.post"})," (enter ",(0,s.jsx)("code",{children:"post"})," here)"]}),(0,s.jsxs)("li",{children:["The current editing field is ",(0,s.jsx)("code",{children:"post.categories"}),", and the relation schema is ",(0,s.jsx)("code",{children:"category"})," (m2m), then the relation field is ",(0,s.jsx)("code",{children:"category.posts"})," (enter ",(0,s.jsx)("code",{children:"posts"})," here)"]})]})]}),children:(0,s.jsx)("span",{className:"mr-1",children:"Relation Field"})})}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{autoComplete:"auto",...l,readOnly:!!(null==n?void 0:n.name)})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"relation.owner",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{className:"flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm",children:[(0,s.jsxs)("div",{className:"space-y-0.5",children:[(0,s.jsx)(i.lX,{children:"Owner"}),(0,s.jsx)(i.pf,{children:"Is the relation field the owner of the relation?"})]}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(I.r,{name:"relation.owner",checked:l.value,onCheckedChange:l.onChange,disabled:!!(null==n?void 0:n.name),"aria-readonly":!!(null==n?void 0:n.name)})})]})}})]})};var F=n(53171);let O=e=>{var l;let{form:n,editingField:a}=e,r=n.watch("type"),t=n.watch("renderer.class"),o=(0,F.gc)(r,null!=a?a:p.bO),d=null!==(l=o.find(e=>e.class===t))&&void 0!==l?l:o[0];return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(i.Wi,{control:n.control,name:"server_name",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{...l,type:"hidden"})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:n.control,name:"type",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Type"}),(0,s.jsxs)(w.Ph,{onValueChange:l.onChange,defaultValue:l.value,name:"type",children:[(0,s.jsx)(i.NI,{children:(0,s.jsx)(w.i4,{children:(0,s.jsx)(w.ki,{placeholder:"Select a type"})})}),(0,s.jsx)(w.Bw,{children:(0,s.jsxs)(y.x,{className:"h-72",children:[(0,s.jsxs)(w.DI,{children:[(0,s.jsx)(w.n5,{children:"Common"}),(0,s.jsx)(w.Ql,{value:"string",children:"Short text"}),(0,s.jsx)(w.Ql,{value:"text",children:"Long text"}),(0,s.jsx)(w.Ql,{value:"bool",children:"Boolean"}),(0,s.jsx)(w.Ql,{value:"int64",children:"Int"}),(0,s.jsx)(w.Ql,{value:"float64",children:"Float"}),(0,s.jsx)(w.Ql,{value:"file",children:"Media"}),(0,s.jsx)(w.Ql,{value:"relation",children:"Relation"})]}),(0,s.jsx)(w.U$,{}),(0,s.jsxs)(w.DI,{children:[(0,s.jsx)(w.n5,{children:"Complex"}),(0,s.jsx)(w.Ql,{value:"bytes",children:"Bytes"}),(0,s.jsx)(w.Ql,{value:"json",children:"Json"}),(0,s.jsx)(w.Ql,{value:"uuid",children:"UUID"})]}),(0,s.jsx)(w.U$,{}),(0,s.jsxs)(w.DI,{children:[(0,s.jsx)(w.n5,{children:"Advanced"}),(0,s.jsx)(w.Ql,{value:"enum",children:"Enum"}),(0,s.jsx)(w.Ql,{value:"time",children:"Time"})]})]})})]}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:n.control,name:"name",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Field name"}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{autoComplete:"auto",placeholder:"age, address...",...l})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:n.control,name:"label",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Field label"}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{placeholder:"Age, Address...",...l})}),(0,s.jsx)(i.zG,{})]})}}),"file"===r&&(0,s.jsx)(i.Wi,{control:n.control,name:"multiple",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{className:"flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm",children:[(0,s.jsxs)("div",{className:"space-y-0.5",children:[(0,s.jsx)(i.lX,{children:"Multiple"}),(0,s.jsx)(i.pf,{children:"Allow multiple values"})]}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(I.r,{name:"unique",checked:l.value,onCheckedChange:l.onChange,"aria-readonly":!0})})]})}}),"enum"===r&&(0,s.jsx)($,{form:n}),"relation"===r&&(0,s.jsx)(z,{form:n,editingField:a}),(0,s.jsx)(S.Z,{className:"my-4"}),(0,s.jsx)(i.Wi,{control:n.control,name:"renderer.class",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Renderer"}),(0,s.jsxs)(w.Ph,{onValueChange:l.onChange,defaultValue:l.value,name:"renderer.class",children:[(0,s.jsx)(i.NI,{children:(0,s.jsx)(w.i4,{children:(0,s.jsx)(w.ki,{placeholder:"Select a renderer"})})}),(0,s.jsx)(w.Bw,{children:o.map(e=>(0,s.jsx)(w.Ql,{value:e.class,children:e.class},e.class))})]}),(0,s.jsx)(i.zG,{})]})}}),d?d.renderSettings(n):null]})},R=e=>{let{form:l}=e;return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(i.Wi,{control:l.control,name:"unique",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{className:"flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm",children:[(0,s.jsxs)("div",{className:"space-y-0.5",children:[(0,s.jsx)(i.lX,{children:"Unique"}),(0,s.jsx)(i.pf,{children:"Prevent duplicate values"})]}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(I.r,{name:"unique",checked:l.value,onCheckedChange:l.onChange,"aria-readonly":!0})})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"optional",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{className:"flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm",children:[(0,s.jsxs)("div",{className:"space-y-0.5",children:[(0,s.jsx)(i.lX,{children:"Optional"}),(0,s.jsx)(i.pf,{children:"Allow null values"})]}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(I.r,{name:"optional",checked:l.value,onCheckedChange:l.onChange,"aria-readonly":!0})})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"sortable",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{className:"flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm",children:[(0,s.jsxs)("div",{className:"space-y-0.5",children:[(0,s.jsx)(i.lX,{children:"Sortable"}),(0,s.jsx)(i.pf,{children:"Allow sorting"})]}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(I.r,{name:"sortable",checked:l.value,onCheckedChange:l.onChange,"aria-readonly":!0})})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"filterable",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{className:"flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm",children:[(0,s.jsxs)("div",{className:"space-y-0.5",children:[(0,s.jsx)(i.lX,{children:"Filterable"}),(0,s.jsx)(i.pf,{children:"Allow filtering"})]}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(I.r,{name:"filterable",checked:l.value,onCheckedChange:l.onChange,"aria-readonly":!0})})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"db.increment",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{className:"flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm",children:[(0,s.jsxs)("div",{className:"space-y-0.5",children:[(0,s.jsx)(i.lX,{children:"Increment"}),(0,s.jsx)(i.pf,{children:"Auto increment the value"})]}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(I.r,{name:"db.increment",checked:l.value,onCheckedChange:l.onChange,"aria-readonly":!0})})]})}})]})},q=e=>{let{form:l}=e,{register:n}=l;return(0,s.jsxs)(s.Fragment,{children:[(0,s.jsx)(i.Wi,{control:l.control,name:"default",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Default"}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{...l,...n("default",{setValueAs:e=>""===e?null:e})})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"size",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Size"}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{type:"number",...l})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"db.attr",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Attribute"}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{...l})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"db.collation",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Collation"}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{...l})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:l.control,name:"db.key",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:"Key"}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{...l})}),(0,s.jsx)(i.zG,{})]})}})]})};var E=n(97081),J=n(52235);function W(e){let{open:l,editingField:n,existingFields:a,onSave:r,onClose:o}=e,d=(0,j.cI)({resolver:(0,u.F)(p.gU),values:{...null!=n?n:p.bO},defaultValues:{...null!=n?n:p.bO}}),c=d.watch("name");(0,f.useEffect)(()=>{(null==n?void 0:n.name)||!c||d.setValue("label",(0,E.rV)(c))},[c]);let m=e=>{if(!(null==n?void 0:n.name)&&(null==a?void 0:a.find(l=>l.name===e.name))){d.setError("name",{type:"manual",message:"Field name already exists."});return}"relation"!==e.type&&delete e.relation,"enum"!==e.type&&delete e.enums,e.db&&0!==Object.keys(e.db).length||delete e.db,""===e.default&&delete e.default,e.optional||delete e.optional,null==r||r(e)};return(0,s.jsx)(v.yo,{open:l,onOpenChange:e=>{e||null==o||o()},children:(0,s.jsxs)(v.ue,{className:"lg:max-w-screen-lg overflow-y-auto max-h-screen py-0 w-full max-w-full field-edit-sheet sm:w-full sm:max-w-full md:w-3/4 md:max-w-3/4",onInteractOutside:e=>{e.preventDefault(),e.stopPropagation()},children:[(0,s.jsxs)(v.Tu,{className:"sticky top-0 z-10 bg-white py-5 text-left",children:[(0,s.jsx)(v.bC,{children:(null==n?void 0:n.name)?"Edit field: ".concat(n.name):"Create new field"}),(0,s.jsx)(v.Ei,{children:(null==n?void 0:n.name)?"Edit the field details below.":"Enter the field details below."}),(0,s.jsx)("button",{type:"button",onClick:o,className:"absolute top-2 right-2",children:(0,s.jsx)(J.Z,{size:20})})]}),(0,s.jsx)(i.l0,{...d,children:(0,s.jsx)("form",{onSubmit:d.handleSubmit(m),children:(0,s.jsx)("div",{className:"grid gap-4 py-4",children:(0,s.jsxs)(g.mQ,{defaultValue:"common",className:"w-full",children:[(0,s.jsxs)(g.dr,{className:"grid w-full grid-cols-3",children:[(0,s.jsx)(g.SP,{value:"common",children:"Common"}),(0,s.jsx)(g.SP,{value:"database",children:"Database"}),(0,s.jsx)(g.SP,{value:"advance",children:"Advance"})]}),(0,s.jsx)(g.nU,{value:"common",children:(0,s.jsx)("div",{className:"grid gap-4 py-4",children:(0,s.jsx)(O,{form:d,editingField:n})})}),(0,s.jsx)(g.nU,{value:"database",children:(0,s.jsx)("div",{className:"grid gap-4 py-4",children:(0,s.jsx)(q,{form:d})})}),(0,s.jsx)(g.nU,{value:"advance",children:(0,s.jsx)("div",{className:"grid gap-4 py-4",children:(0,s.jsx)(R,{form:d})})})]})})})}),(0,s.jsx)(v.FF,{className:"sticky bottom-0 z-10 bg-white py-5",children:(0,s.jsx)(t.z,{type:"submit",onClick:()=>{d.handleSubmit(m)()},children:"Save field"})})]})})}var X=n(12647);let P=e=>{var l,n,a;let{form:r,fields:t,editingSchema:o}=e;return(0,s.jsxs)(s.Fragment,{children:[r.formState.errors.fields&&(0,s.jsx)(_.T,{description:null!==(a=r.formState.errors.fields.message)&&void 0!==a?a:null===(n=r.formState.errors.fields)||void 0===n?void 0:null===(l=n.root)||void 0===l?void 0:l.message}),(0,s.jsx)(i.Wi,{control:r.control,name:"name",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{className:"flex",children:(0,s.jsx)(k.u,{tip:"This is the name of your schema.",icon:!0,children:(0,s.jsx)("span",{className:"mr-1",children:"Name"})})}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{...l,autoComplete:"auto",placeholder:"Schema name"})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:r.control,name:"namespace",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{className:"flex",children:(0,s.jsx)(k.u,{icon:!0,tip:"This is the namespace of your schema. It will be used to generate the database table name and API endpoints.",children:(0,s.jsx)("span",{className:"mr-1",children:"Namespace"})})}),(0,s.jsx)(i.NI,{children:(0,s.jsx)(b.I,{...l,placeholder:"Schema namespace",className:(null==o?void 0:o.name)?"read-only:bg-gray-100":""})}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:r.control,name:"label_field",render:e=>{let{field:l}=e;return(0,s.jsxs)(i.xJ,{children:[(0,s.jsx)(i.lX,{children:(0,s.jsx)(k.u,{icon:!0,tip:"This is the namespace of your schema. It will be used to generate the database table name and API endpoints",children:(0,s.jsx)("span",{className:"mr-1",children:"Label Field"})})}),(0,s.jsxs)(w.Ph,{onValueChange:l.onChange,value:l.value,name:"label_field",children:[(0,s.jsx)(i.NI,{children:(0,s.jsx)(w.i4,{children:(0,s.jsx)(w.ki,{placeholder:"Select a label field"})})}),(0,s.jsx)(w.Bw,{children:(0,s.jsx)(y.x,{className:"h-72",children:t.map((e,l)=>(0,s.jsxs)(w.Ql,{value:e.name,children:[e.name," - ",e.label]},e.name))})})]}),(0,s.jsx)(i.zG,{})]})}}),(0,s.jsx)(i.Wi,{control:r.control,name:"disable_timestamp",render:e=>{let{field:l}=e;return(0,s.jsx)(i.xJ,{children:(0,s.jsx)(i.NI,{children:(0,s.jsxs)("div",{className:"flex items-center space-x-2",children:[(0,s.jsx)(I.r,{checked:l.value,onCheckedChange:l.onChange,name:"disable_timestamp",id:"disable_timestamp","aria-readonly":!0}),(0,s.jsx)(X._,{htmlFor:"disable_timestamp",children:"Disable timestamps"})]})})})}})]})};var Z=n(5887),G=n(77731),V=n(66806),Q=n(33277);let A=e=>{var l,n,a;let{form:r,fields:i,setEditingField:o,removeField:d}=e,c=(0,s.jsx)(t.z,{type:"button",variant:"default",size:"sm",icon:(0,s.jsx)(N.Z,{size:16}),onClick:()=>o(p.bO),children:"New Field"});return(0,s.jsxs)("div",{className:"space-y-4",children:[r.formState.errors.fields&&(0,s.jsx)(_.T,{description:null!==(a=r.formState.errors.fields.message)&&void 0!==a?a:null===(n=r.formState.errors.fields)||void 0===n?void 0:null===(l=n.root)||void 0===l?void 0:l.message}),c,(0,s.jsxs)(Z.iA,{children:[(0,s.jsx)(Z.xD,{children:(0,s.jsxs)(Z.SC,{children:[(0,s.jsx)(Z.ss,{children:"Label"}),(0,s.jsx)(Z.ss,{children:"Name"}),(0,s.jsx)(Z.ss,{children:"Type"}),(0,s.jsx)(Z.ss,{children:"System"}),(0,s.jsx)(Z.ss,{children:"Optional"}),(0,s.jsx)(Z.ss,{className:"text-right",children:"Actions"})]})}),(0,s.jsx)(Z.RM,{children:i.map((e,l)=>{var n;return(0,s.jsxs)(Z.SC,{children:[(0,s.jsx)(Z.pj,{className:"font-medium truncate",children:e.label}),(0,s.jsx)(Z.pj,{className:"font-medium truncate",children:e.name}),(0,s.jsxs)(Z.pj,{className:"truncate",children:[e.type,"relation"===e.type&&(0,s.jsx)(Q.C,{variant:"secondary",className:"ml-2",children:null==e?void 0:null===(n=e.relation)||void 0===n?void 0:n.type})]}),(0,s.jsx)(Z.pj,{children:e.is_system_field?(0,s.jsx)(Q.C,{variant:"secondary",children:"System"}):null}),(0,s.jsx)(Z.pj,{children:e.optional?null:(0,s.jsx)(Q.C,{variant:"destructive",children:"Required"})}),(0,s.jsx)(Z.pj,{className:"text-right",children:!e.is_system_field&&(0,s.jsxs)(s.Fragment,{children:[(0,s.jsxs)("button",{type:"button",className:"text-sm inline-flex flex-row items-center gap-1 hover:underline pr-3",onClick:()=>o(e),children:[(0,s.jsx)(G.Z,{size:12})," Edit"]}),(0,s.jsxs)("button",{type:"button",className:"text-sm inline-flex flex-row items-center gap-1 text-red-800 hover:underline",onClick:()=>{window.confirm("Are you sure you want to delete this field?")&&d(l)},children:[(0,s.jsx)(V.Z,{size:12}),"Delete"]})]})})]},e.name)})})]}),i.length?c:null]})},L=e=>{let l=(0,m.useRouter)(),{reloadAppConfig:n}=(0,f.useContext)(o.Il),{editingSchema:a}=e,[r,h]=(0,f.useState)(),v=(0,j.cI)({resolver:(0,u.F)(p.Pl),defaultValues:null!=a?a:p.IG,mode:"onChange"}),g=v.watch("name"),y=v.watch("label_field"),b=v.watch("fields"),{fields:w,append:N,update:$,remove:C}=(0,j.Dq)({name:"fields",control:v.control});(0,f.useEffect)(()=>{v.reset(null!=a?a:p.IG)},[null==a?void 0:a.name]),(0,f.useEffect)(()=>{(null==a?void 0:a.name)||!g||v.setValue("namespace",x()(g.trim()))},[g]),(0,f.useEffect)(()=>{if(b&&b.length>0&&!b.some(e=>e.name===y)){let e=b.find(e=>"string"===e.type);e&&v.setValue("label_field",e.name)}},[b,y]);let I=async e=>{let s=[],r=!1;for(let l of e.fields)l.server_name&&l.name!==l.server_name&&s.push({from:l.server_name,to:l.name}),l.name===e.label_field&&(r=!0);if(!r){d.h4.error("Label field is invalid."),v.setValue("label_field",""),v.setError("label_field",{type:"manual",message:"Label field is invalid."});return}try{var t;let r=await (0,c.uX)(e,null!==(t=null==a?void 0:a.name)&&void 0!==t?t:"",s);d.h4.success("Schema ".concat(r.name," saved successfully.")),l.push("/schemas/edit?schema=".concat(r.name)),v.reset(r),n()}catch(e){}};return(0,s.jsxs)("div",{className:"space-y-5",children:[(0,s.jsx)(i.l0,{...v,children:(0,s.jsxs)("form",{className:"grid gap-8 md:grid-cols-2 lg:grid-cols-3",onSubmit:v.handleSubmit(I),children:[(0,s.jsx)("div",{className:"relative flex-col items-start gap-8 md:flex",children:(0,s.jsxs)("fieldset",{className:"sticky top-5 grid w-full gap-5 rounded-lg border p-4",children:[(0,s.jsx)("legend",{className:"-ml-1 px-1 text-sm font-medium",children:"Schema"}),(0,s.jsx)(P,{form:v,fields:w,editingSchema:a}),(0,s.jsx)(t.z,{type:"submit",children:"Save"})]})}),(0,s.jsx)("div",{className:"flex-col relative flex rounded-xl lg:col-span-2",children:(0,s.jsxs)("fieldset",{className:"grid w-full gap-5 rounded-lg border p-4",children:[(0,s.jsx)("legend",{className:"-ml-1 px-1 text-sm font-medium",children:"Fields"}),(0,s.jsx)(A,{form:v,fields:w,setEditingField:h,removeField:C})]})})]})}),r?(0,s.jsx)(W,{open:!!r,editingField:r,existingFields:w,onClose:()=>h(void 0),onSave:e=>{if(null==r?void 0:r.name){let l=w.findIndex(e=>e.name===r.name);l>=0&&$(l,e)}else N(e);h(void 0)}}):null]})};function D(){let e=(0,m.useSearchParams)().get("schema"),{data:l,isLoading:n,error:t}=(0,a.a)({queryKey:["schema",e],queryFn:()=>(0,c.J1)(null!=e?e:""),retry:!1,refetchOnWindowFocus:!1});return((0,f.useEffect)(()=>{let n=(null==l?void 0:l.name)?"Edit schema: ".concat(l.name):"Create new schema",s=(null==l?void 0:l.name)?"Edit schema ".concat(l.name," to change the structure of your data."):"Create a new schema for your data.";return(0,r.D8)({title:n,description:s,breadcrumbs:[{name:"Schema",path:"/schemas"},{name:e?"Edit schema ".concat(e):"New schema",path:"/schemas/edit?schema="+e}]}),r.D8},[e,l]),n)?(0,s.jsx)(C.g,{}):t?(0,s.jsx)(_.T,{error:t}):(0,s.jsx)(L,{editingSchema:l})}},99773:function(e,l,n){"use strict";n.d(l,{IG:function(){return i},Pl:function(){return r},bO:function(){return t},gU:function(){return a}});var s=n(30248);let a=s.Ry({name:s.Z_().trim().min(1,{message:"Field name is required"}),server_name:s.Z_().optional(),label:s.Z_().min(1,{message:"Field label is required"}),type:s.Km(["bool","time","json","uuid","bytes","enum","string","text","int","int8","int16","int32","int64","uint","uint8","uint16","uint32","uint64","float32","float64","relation","file"]),multiple:s.O7().optional(),size:s.oQ.number().optional(),unique:s.O7().optional(),optional:s.O7().optional(),default:s.Yj().nullable(),sortable:s.O7().optional(),filterable:s.O7().optional(),renderer:s.Ry({class:s.Z_().optional(),settings:s.IM(s.Z_(),s.Yj()).optional()}).optional(),enums:s.IX(s.Ry({value:s.Z_().min(1,{message:"Enum value is required"}),label:s.Z_().min(1,{message:"Enum label is required"})})).optional(),relation:s.Ry({schema:s.Z_(),field:s.Z_(),type:s.Km(["o2o","o2m","m2m"]),owner:s.O7().optional(),fk_columns:s.IM(s.Z_(),s.Z_()).optional().nullable(),junction_table:s.Z_().optional(),optional:s.O7().optional()}).optional(),db:s.Ry({attr:s.Z_().optional(),collation:s.Z_().optional(),increment:s.O7().optional(),key:s.Z_().optional()}).nullable().optional(),is_system_field:s.O7().optional()}).superRefine((e,l)=>{var n;return"enum"!==e.type||null!=e&&null!==(n=e.enums)&&void 0!==n&&!!n.length||l.addIssue({code:s.NL.custom,message:"Enums are required for enum type",path:["type"]})}).superRefine((e,l)=>{if("relation"===e.type){var n,a,r,t,i,o;(null==e?void 0:null===(n=e.relation)||void 0===n?void 0:n.type)&&(null==e?void 0:null===(a=e.relation)||void 0===a?void 0:a.schema)&&(null==e?void 0:null===(r=e.relation)||void 0===r?void 0:r.field)||l.addIssue({code:s.NL.custom,message:"Relation type, schema, field is required for relation type",path:["type"]}),(null==e?void 0:null===(t=e.relation)||void 0===t?void 0:t.type)||l.addIssue({code:s.NL.custom,message:"Relation type, schema, field is required for relation type",path:["relation.type"]}),(null==e?void 0:null===(i=e.relation)||void 0===i?void 0:i.schema)||l.addIssue({code:s.NL.custom,message:"Relation schema is required for relation type",path:["relation.schema"]}),(null==e?void 0:null===(o=e.relation)||void 0===o?void 0:o.field)||l.addIssue({code:s.NL.custom,message:"Relation field is required for relation type",path:["relation.field"]})}return!0}),r=s.Ry({name:s.Z_().trim().min(1,{message:"Schema name is required"}).regex(/^[a-zA-Z]\w*$/,{message:"Schema name should start with an alphabet and contain only alphabets, numbers, or underscores"}),namespace:s.Z_().trim().min(1,{message:"Schema namespace is required"}),label_field:s.Z_().trim().min(1,{message:"Schema label field is required"}),disable_timestamp:s.O7(),is_system_schema:s.O7().optional(),fields:s.IX(a).min(1,{message:"At least one field is required"})}).refine(e=>{var l;let n=null==e?void 0:null===(l=e.fields)||void 0===l?void 0:l.map(e=>e.name),s=[...new Set(n)];return n.length===s.length},{message:"Field names must be unique"}),t={name:"",server_name:"",label:"",type:"string",enums:[],default:"",size:0,multiple:!1,db:{attr:"",collation:"",increment:!1,key:""},renderer:{class:"",settings:{}},relation:{schema:"",field:"",type:"o2o",owner:!1,fk_columns:{},junction_table:"",optional:!1},unique:!1,optional:!0,sortable:!0,filterable:!1,is_system_field:!1},i={name:"",namespace:"",label_field:"",disable_timestamp:!1,fields:[]}},50326:function(e,l,n){"use strict";n.d(l,{Ei:function(){return v},FF:function(){return j},Tu:function(){return f},aM:function(){return c},bC:function(){return p},ue:function(){return x},yo:function(){return d}});var s=n(57437),a=n(2265),r=n(72936),t=n(62177),i=n(57742),o=n(22169);let d=r.fC,c=r.xz;r.x8;let u=r.h_,m=a.forwardRef((e,l)=>{let{className:n,...a}=e;return(0,s.jsx)(r.aV,{className:(0,o.cn)("fixed inset-0 z-50 bg-black/80  data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",n),...a,ref:l})});m.displayName=r.aV.displayName;let h=(0,i.j)("fixed z-50 gap-4 bg-background p-6 shadow-lg transition ease-in-out data-[state=closed]:duration-300 data-[state=open]:duration-500 data-[state=open]:animate-in data-[state=closed]:animate-out",{variants:{side:{top:"inset-x-0 top-0 border-b data-[state=closed]:slide-out-to-top data-[state=open]:slide-in-from-top",bottom:"inset-x-0 bottom-0 border-t data-[state=closed]:slide-out-to-bottom data-[state=open]:slide-in-from-bottom",left:"inset-y-0 left-0 h-full w-3/4 border-r data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left sm:max-w-sm",right:"inset-y-0 right-0 h-full w-3/4 border-l data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right sm:max-w-sm"}},defaultVariants:{side:"right"}}),x=a.forwardRef((e,l)=>{let{side:n="right",className:a,children:i,...d}=e;return(0,s.jsxs)(u,{children:[(0,s.jsx)(m,{}),(0,s.jsxs)(r.VY,{ref:l,className:(0,o.cn)(h({side:n}),a),...d,children:[(0,s.jsxs)(r.x8,{className:"absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-secondary",children:[(0,s.jsx)(t.Pxu,{className:"h-4 w-4"}),(0,s.jsx)("span",{className:"sr-only",children:"Close"})]}),i]})]})});x.displayName=r.VY.displayName;let f=e=>{let{className:l,...n}=e;return(0,s.jsx)("div",{className:(0,o.cn)("flex flex-col space-y-2 text-center sm:text-left",l),...n})};f.displayName="SheetHeader";let j=e=>{let{className:l,...n}=e;return(0,s.jsx)("div",{className:(0,o.cn)("flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2",l),...n})};j.displayName="SheetFooter";let p=a.forwardRef((e,l)=>{let{className:n,...a}=e;return(0,s.jsx)(r.Dx,{ref:l,className:(0,o.cn)("text-lg font-semibold text-foreground",n),...a})});p.displayName=r.Dx.displayName;let v=a.forwardRef((e,l)=>{let{className:n,...a}=e;return(0,s.jsx)(r.dk,{ref:l,className:(0,o.cn)("text-sm text-muted-foreground",n),...a})});v.displayName=r.dk.displayName},96304:function(e,l,n){"use strict";n.d(l,{J1:function(){return t},J2:function(){return i},cQ:function(){return a},uV:function(){return d},uX:function(){return r},yR:function(){return o}});var s=n(84971);let a=async()=>(await (0,s.dX)("/schema")).filter(e=>!e.is_junction_schema),r=async(e,l,n)=>{var a,r;let t=null;return(null==(t=l?await (0,s.qb)("/schema/".concat(l),{schema:e,rename_fields:null!=n?n:[]}):await (0,s.SO)("/schema",e))?void 0:null===(a=t.fields)||void 0===a?void 0:a.length)&&(t.fields=(null!==(r=null==t?void 0:t.fields)&&void 0!==r?r:[]).map(e=>(e.server_name=e.name,e))),t},t=async e=>{var l,n;if(!e)return null;let a=await (0,s.dX)("/schema/".concat(e));return(null==a?void 0:null===(l=a.fields)||void 0===l?void 0:l.length)&&(a.fields=(null!==(n=null==a?void 0:a.fields)&&void 0!==n?n:[]).map(e=>(e.server_name=e.name,e))),a},i=async e=>await (0,s.SO)("/schema/export",e),o=async e=>{let l=new FormData;return e.forEach(e=>l.append("file",e)),await (0,s.SO)("/schema/import",l,{headers:{"Content-Type":"multipart/form-data"}})},d=e=>(0,s.HG)("/schema/".concat(e))}},function(e){e.O(0,[310,572,902,637,792,872,571,32,732,532,152,998,160,147,255,778,564,408,547,117,271,650,971,69,744],function(){return e(e.s=2071)}),_N_E=e.O()}]);