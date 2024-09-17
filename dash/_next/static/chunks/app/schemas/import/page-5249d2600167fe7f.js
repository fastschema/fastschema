(self.webpackChunk_N_E=self.webpackChunk_N_E||[]).push([[934],{31873:function(e,r,t){Promise.resolve().then(t.bind(t,81626))},47907:function(e,r,t){"use strict";var n=t(15313);t.o(n,"ReadonlyURLSearchParams")&&t.d(r,{ReadonlyURLSearchParams:function(){return n.ReadonlyURLSearchParams}}),t.o(n,"usePathname")&&t.d(r,{usePathname:function(){return n.usePathname}}),t.o(n,"useRouter")&&t.d(r,{useRouter:function(){return n.useRouter}}),t.o(n,"useSearchParams")&&t.d(r,{useSearchParams:function(){return n.useSearchParams}})},81626:function(e,r,t){"use strict";t.r(r),t.d(r,{default:function(){return p}});var n=t(57437),a=t(2265),l=t(34560),s=t(94509),i=t(54739),c=t(22169),o=t(57654),u=t(96304),d=t(79984),f=t(27453),m=t(47907);let h=e=>{let{uploadContainerClass:r,minimal:t}=e,{getRootProps:l,getInputProps:h,isDragActive:p,acceptedFiles:y}=(0,i.uI)(),{reloadAppConfig:x}=(0,a.useContext)(d.Il),b=(0,m.useRouter)();return(0,a.useEffect)(()=>{y.length&&(async()=>{try{await (0,u.yR)(y),x(),b.push("/schemas")}catch(e){f.h4.error(e.message)}})()},[y]),(0,n.jsx)("div",{className:"space-y-5",children:t?(0,n.jsx)("div",{className:"flex flex-row space-x-2 align-middle items-center",children:(0,n.jsxs)("div",{...l(),children:[(0,n.jsx)("input",{...h(),type:"file",className:"hidden"}),(0,n.jsxs)(o.z,{size:"sm",variant:"outline",children:[(0,n.jsx)(s.Z,{className:"mr-2 h-4 w-4"}),(0,n.jsx)("span",{children:"Upload"})]})]})}):(0,n.jsxs)("div",{...l(),className:(0,c.cn)("flex flex-col items-center justify-center w-full h-64 border-2 border-gray-300 border-dashed rounded-lg cursor-pointer bg-gray-50 dark:hover:bg-bray-800 dark:bg-gray-700 hover:bg-gray-100 dark:border-gray-600 dark:hover:border-gray-500 dark:hover:bg-gray-600",r),children:[(0,n.jsxs)("div",{className:"flex flex-col items-center justify-center pt-5 pb-6",children:[(0,n.jsx)(s.Z,{className:"w-8 h-8 mb-4 text-gray-500 dark:text-gray-400"}),(0,n.jsx)("p",{className:"mb-2 text-sm text-gray-500 dark:text-gray-400",children:p?"Drop the files here ...":"Drag and drop some files here, or click to select files"}),(0,n.jsx)("p",{className:"text-xs text-gray-500 dark:text-gray-400",children:"JSON"})]}),(0,n.jsx)("input",{...h(),id:"dropzone-file",type:"file",className:"hidden"})]})})};function p(){return(0,a.useEffect)(()=>((0,l.D8)({title:"Import Schemas",description:"Import schemas.",breadcrumbs:[{name:"Schema",path:"/schemas"},{name:"Import",path:"/schemas/import"}]}),l.D8),[]),(0,n.jsx)(h,{})}},34560:function(e,r,t){"use strict";t.d(r,{Sc:function(){return g},D8:function(){return v}});var n=t(57437),a=t(8792),l=t(13571),s=t.n(l),i=t(51919),c=t(2265),o=t(62177),u=t(59143),d=t(22169);let f=c.forwardRef((e,r)=>{let{...t}=e;return(0,n.jsx)("nav",{ref:r,"aria-label":"breadcrumb",...t})});f.displayName="Breadcrumb";let m=c.forwardRef((e,r)=>{let{className:t,...a}=e;return(0,n.jsx)("ol",{ref:r,className:(0,d.cn)("flex flex-wrap items-center gap-1.5 break-words text-sm text-muted-foreground sm:gap-2.5",t),...a})});m.displayName="BreadcrumbList";let h=c.forwardRef((e,r)=>{let{className:t,...a}=e;return(0,n.jsx)("li",{ref:r,className:(0,d.cn)("inline-flex items-center gap-1.5",t),...a})});h.displayName="BreadcrumbItem";let p=c.forwardRef((e,r)=>{let{asChild:t,className:a,...l}=e,s=t?u.g7:"a";return(0,n.jsx)(s,{ref:r,className:(0,d.cn)("transition-colors hover:text-foreground",a),...l})});p.displayName="BreadcrumbLink";let y=c.forwardRef((e,r)=>{let{className:t,...a}=e;return(0,n.jsx)("span",{ref:r,role:"link","aria-disabled":"true","aria-current":"page",className:(0,d.cn)("font-normal text-foreground",t),...a})});y.displayName="BreadcrumbPage";let x=e=>{let{children:r,className:t,...a}=e;return(0,n.jsx)("li",{role:"presentation","aria-hidden":"true",className:(0,d.cn)("[&>svg]:size-3.5",t),...a,children:null!=r?r:(0,n.jsx)(o.XCv,{})})};x.displayName="BreadcrumbSeparator";let b={title:"",description:"",breadcrumbs:[],actions:[]},v=e=>{i.ZP.dispatch("pageInfo",null!=e?e:[])},g=()=>{var e,r;let[t,l]=(0,c.useState)(b);(0,c.useEffect)(()=>{i.ZP.on("pageInfo",e=>{l(null!=e?e:b)})},[]);let o=(null==t?void 0:null===(e=t.breadcrumbs)||void 0===e?void 0:e.length)?[{name:"Dash",path:"/"},...null!==(r=t.breadcrumbs)&&void 0!==r?r:[]]:[],u=o.length;return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)("title",{children:t.title}),(0,n.jsx)(f,{children:(0,n.jsx)(m,{children:s()(o,e=>e.path).map((e,r)=>{let t=r===u-1;return(0,n.jsxs)(c.Fragment,{children:[(0,n.jsx)(h,{children:t?(0,n.jsx)(y,{children:e.name}):(0,n.jsx)(p,{asChild:!0,children:(0,n.jsx)(a.default,{href:e.path,children:e.name})})}),!t&&(0,n.jsx)(x,{})]},e.path)})})}),(t.title||t.description)&&(0,n.jsxs)("div",{className:"flex items-center justify-between space-y-2 w-full",children:[(0,n.jsxs)("div",{children:[t.title&&(0,n.jsx)("h1",{className:"text-lg font-semibold md:text-2xl",children:t.title}),t.description&&(0,n.jsx)("p",{className:"text-muted-foreground",children:t.description})]}),(0,n.jsx)("div",{className:"ml-auto mr-4 flex gap-2",children:t.actions})]})]})}},96304:function(e,r,t){"use strict";t.d(r,{J1:function(){return s},J2:function(){return i},cQ:function(){return a},uV:function(){return o},uX:function(){return l},yR:function(){return c}});var n=t(84971);let a=async()=>(await (0,n.dX)("/schema")).filter(e=>!e.is_junction_schema),l=async(e,r,t)=>{var a,l;let s=null;return(null==(s=r?await (0,n.qb)("/schema/".concat(r),{schema:e,rename_fields:null!=t?t:[]}):await (0,n.SO)("/schema",e))?void 0:null===(a=s.fields)||void 0===a?void 0:a.length)&&(s.fields=(null!==(l=null==s?void 0:s.fields)&&void 0!==l?l:[]).map(e=>(e.server_name=e.name,e))),s},s=async e=>{var r,t;if(!e)return null;let a=await (0,n.dX)("/schema/".concat(e));return(null==a?void 0:null===(r=a.fields)||void 0===r?void 0:r.length)&&(a.fields=(null!==(t=null==a?void 0:a.fields)&&void 0!==t?t:[]).map(e=>(e.server_name=e.name,e))),a},i=async e=>await (0,n.SO)("/schema/export",e),c=async e=>{let r=new FormData;return e.forEach(e=>r.append("file",e)),await (0,n.SO)("/schema/import",r,{headers:{"Content-Type":"multipart/form-data"}})},o=e=>(0,n.HG)("/schema/".concat(e))},69703:function(e,r,t){"use strict";t.d(r,{CR:function(){return i},Jh:function(){return s},_T:function(){return a},ev:function(){return c},mG:function(){return l},pi:function(){return n}});var n=function(){return(n=Object.assign||function(e){for(var r,t=1,n=arguments.length;t<n;t++)for(var a in r=arguments[t])Object.prototype.hasOwnProperty.call(r,a)&&(e[a]=r[a]);return e}).apply(this,arguments)};function a(e,r){var t={};for(var n in e)Object.prototype.hasOwnProperty.call(e,n)&&0>r.indexOf(n)&&(t[n]=e[n]);if(null!=e&&"function"==typeof Object.getOwnPropertySymbols)for(var a=0,n=Object.getOwnPropertySymbols(e);a<n.length;a++)0>r.indexOf(n[a])&&Object.prototype.propertyIsEnumerable.call(e,n[a])&&(t[n[a]]=e[n[a]]);return t}function l(e,r,t,n){return new(t||(t=Promise))(function(a,l){function s(e){try{c(n.next(e))}catch(e){l(e)}}function i(e){try{c(n.throw(e))}catch(e){l(e)}}function c(e){var r;e.done?a(e.value):((r=e.value)instanceof t?r:new t(function(e){e(r)})).then(s,i)}c((n=n.apply(e,r||[])).next())})}function s(e,r){var t,n,a,l,s={label:0,sent:function(){if(1&a[0])throw a[1];return a[1]},trys:[],ops:[]};return l={next:i(0),throw:i(1),return:i(2)},"function"==typeof Symbol&&(l[Symbol.iterator]=function(){return this}),l;function i(i){return function(c){return function(i){if(t)throw TypeError("Generator is already executing.");for(;l&&(l=0,i[0]&&(s=0)),s;)try{if(t=1,n&&(a=2&i[0]?n.return:i[0]?n.throw||((a=n.return)&&a.call(n),0):n.next)&&!(a=a.call(n,i[1])).done)return a;switch(n=0,a&&(i=[2&i[0],a.value]),i[0]){case 0:case 1:a=i;break;case 4:return s.label++,{value:i[1],done:!1};case 5:s.label++,n=i[1],i=[0];continue;case 7:i=s.ops.pop(),s.trys.pop();continue;default:if(!(a=(a=s.trys).length>0&&a[a.length-1])&&(6===i[0]||2===i[0])){s=0;continue}if(3===i[0]&&(!a||i[1]>a[0]&&i[1]<a[3])){s.label=i[1];break}if(6===i[0]&&s.label<a[1]){s.label=a[1],a=i;break}if(a&&s.label<a[2]){s.label=a[2],s.ops.push(i);break}a[2]&&s.ops.pop(),s.trys.pop();continue}i=r.call(e,s)}catch(e){i=[6,e],n=0}finally{t=a=0}if(5&i[0])throw i[1];return{value:i[0]?i[1]:void 0,done:!0}}([i,c])}}}function i(e,r){var t="function"==typeof Symbol&&e[Symbol.iterator];if(!t)return e;var n,a,l=t.call(e),s=[];try{for(;(void 0===r||r-- >0)&&!(n=l.next()).done;)s.push(n.value)}catch(e){a={error:e}}finally{try{n&&!n.done&&(t=l.return)&&t.call(l)}finally{if(a)throw a.error}}return s}function c(e,r,t){if(t||2==arguments.length)for(var n,a=0,l=r.length;a<l;a++)!n&&a in r||(n||(n=Array.prototype.slice.call(r,0,a)),n[a]=r[a]);return e.concat(n||Array.prototype.slice.call(r))}"function"==typeof SuppressedError&&SuppressedError}},function(e){e.O(0,[310,637,792,571,532,564,971,69,744],function(){return e(e.s=31873)}),_N_E=e.O()}]);