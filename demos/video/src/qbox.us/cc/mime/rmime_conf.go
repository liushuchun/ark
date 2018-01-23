package mime

var rmime_conf = `
{
	"applicaiton/x-bytecode.python": [
		".pyc"
	],
	"application/andrew-inset": [
		".ez"
	],
	"application/annodex": [
		".anx"
	],
	"application/applixware": [
		".aw"
	],
	"application/arj": [
		".arj"
	],
	"application/atom+xml": [
		".atom"
	],
	"application/atomcat+xml": [
		".atomcat"
	],
	"application/atomserv+xml": [
		".atomsrv"
	],
	"application/atomsvc+xml": [
		".atomsvc"
	],
	"application/base64": [
		".mm",
		".mme"
	],
	"application/bbolin": [
		".lin"
	],
	"application/book": [
		".book",
		".boo"
	],
	"application/cap": [
		".pcap",
		".cap"
	],
	"application/ccxml+xml,": [
		".ccxml"
	],
	"application/cdf": [
		".cdf"
	],
	"application/cdmi-capability": [
		".cdmia"
	],
	"application/cdmi-container": [
		".cdmic"
	],
	"application/cdmi-domain": [
		".cdmid"
	],
	"application/cdmi-object": [
		".cdmio"
	],
	"application/cdmi-queue": [
		".cdmiq"
	],
	"application/clariscad": [
		".ccad"
	],
	"application/cu-seeme": [
		".cu"
	],
	"application/davmount+xml": [
		".davmount"
	],
	"application/drafting": [
		".drw"
	],
	"application/dsptype": [
		".tsp"
	],
	"application/dssc+der": [
		".dssc"
	],
	"application/dssc+xml": [
		".xdssc"
	],
	"application/ecmascript": [
		".es"
	],
	"application/emma+xml": [
		".emma"
	],
	"application/envoy": [
		".evy"
	],
	"application/epub+zip": [
		".epub"
	],
	"application/excel": [
		".xlc",
		".xl",
		".xlw",
		".xlt",
		".xlk",
		".xlv",
		".xla",
		".xld",
		".xlm",
		".xll",
		".xlb"
	],
	"application/exi": [
		".exi"
	],
	"application/font-tdpfr": [
		".pfr"
	],
	"application/fractals": [
		".fif"
	],
	"application/freeloader": [
		".frl"
	],
	"application/gnutar": [
		".tgz"
	],
	"application/groupwise": [
		".vew"
	],
	"application/hta": [
		".hta"
	],
	"application/hyperstudio": [
		".stk"
	],
	"application/i-deas": [
		".unv"
	],
	"application/iges": [
		".iges"
	],
	"application/inf": [
		".inf"
	],
	"application/ipfix": [
		".ipfix"
	],
	"application/java-archive": [
		".jar"
	],
	"application/java-serialized-object": [
		".ser"
	],
	"application/java-vm": [
		".class"
	],
	"application/javascript": [
		".js"
	],
	"application/json": [
		".json"
	],
	"application/lha": [
		".lha"
	],
	"application/lzx": [
		".lzx"
	],
	"application/m3g": [
		".m3g"
	],
	"application/mac-binhex40": [
		".hqx"
	],
	"application/mac-compactpro": [
		".cpt"
	],
	"application/mads+xml": [
		".mads"
	],
	"application/marc": [
		".mrc"
	],
	"application/marcxml+xml": [
		".mrcx"
	],
	"application/mathematica": [
		".nb",
		".ma"
	],
	"application/mathml+xml": [
		".mathml"
	],
	"application/mbedlet": [
		".mbd"
	],
	"application/mbox": [
		".mbox"
	],
	"application/mediaservercontrol+xml": [
		".mscml"
	],
	"application/metalink4+xml": [
		".meta4"
	],
	"application/mets+xml": [
		".mets"
	],
	"application/mime": [
		".aps"
	],
	"application/mods+xml": [
		".mods"
	],
	"application/mp21": [
		".m21"
	],
	"application/mspowerpoint": [
		".pps",
		".ppz",
		".pot"
	],
	"application/msword": [
		".doc",
		".dot",
		".wiz",
		".w6w",
		".word"
	],
	"application/mxf": [
		".mxf"
	],
	"application/netmc": [
		".mcp"
	],
	"application/octet-stream": [
		".ipa",
		".com",
		".o",
		".saveme",
		".zoo",
		".lzh",
		".dump",
		".lhx",
		".a",
		".bin",
		".arc"
	],
	"application/oda": [
		".oda"
	],
	"application/oebps-package+xml": [
		".opf"
	],
	"application/ogg": [
		".ogg",
		".ogx"
	],
	"application/onenote": [
		".one",
		".onetoc2",
		".onetmp",
		".onetoc",
		".onepkg"
	],
	"application/patch-ops-error+xml": [
		".xer"
	],
	"application/pdf": [
		".pdf"
	],
	"application/pgp-keys": [
		".key"
	],
	"application/pgp-signature": [
		".pgp"
	],
	"application/pics-rules": [
		".prf"
	],
	"application/pkcs10": [
		".p10"
	],
	"application/pkcs7-mime": [
		".p7m",
		".p7c"
	],
	"application/pkcs7-signature": [
		".p7s"
	],
	"application/pkcs8": [
		".p8"
	],
	"application/pkix-attr-cert": [
		".ac"
	],
	"application/pkix-cert": [
		".cer",
		".crt"
	],
	"application/pkix-crl": [
		".crl"
	],
	"application/pkix-pkipath": [
		".pkipath"
	],
	"application/pkixcmp": [
		".pki"
	],
	"application/plain": [
		".text"
	],
	"application/pls+xml": [
		".pls"
	],
	"application/postscript": [
		".ai",
		".epsi",
		".eps2",
		".eps",
		".epsf",
		".ps",
		".eps3"
	],
	"application/pro_eng": [
		".part",
		".prt"
	],
	"application/prs.cww": [
		".cww"
	],
	"application/pskc+xml": [
		".pskcxml"
	],
	"application/rdf+xml": [
		".rdf"
	],
	"application/reginfo+xml": [
		".rif"
	],
	"application/relax-ng-compact-syntax": [
		".rnc"
	],
	"application/resource-lists+xml": [
		".rl"
	],
	"application/resource-lists-diff+xml": [
		".rld"
	],
	"application/ringing-tones": [
		".rng"
	],
	"application/rls-services+xml": [
		".rs"
	],
	"application/rsd+xml": [
		".rsd"
	],
	"application/rss+xml": [
		".rss"
	],
	"application/rtf": [
		".rtf"
	],
	"application/sbml+xml": [
		".sbml"
	],
	"application/scvp-cv-request": [
		".scq"
	],
	"application/scvp-cv-response": [
		".scs"
	],
	"application/scvp-vp-request": [
		".spq"
	],
	"application/scvp-vp-response": [
		".spp"
	],
	"application/sdp": [
		".sdp"
	],
	"application/sea": [
		".sea"
	],
	"application/set": [
		".set"
	],
	"application/set-payment-initiation": [
		".setpay"
	],
	"application/set-registration-initiation": [
		".setreg"
	],
	"application/shf+xml": [
		".shf"
	],
	"application/smil": [
		".smil"
	],
	"application/smil+xml": [
		".smi"
	],
	"application/solids": [
		".sol"
	],
	"application/sounder": [
		".sdr"
	],
	"application/sparql-query": [
		".rq"
	],
	"application/sparql-results+xml": [
		".srx"
	],
	"application/srgs": [
		".gram"
	],
	"application/srgs+xml": [
		".grxml"
	],
	"application/sru+xml": [
		".sru"
	],
	"application/ssml+xml": [
		".ssml"
	],
	"application/step": [
		".step",
		".stp"
	],
	"application/streamingmedia": [
		".ssm"
	],
	"application/tei+xml": [
		".tei"
	],
	"application/thraud+xml": [
		".tfi"
	],
	"application/timestamped-data": [
		".tsd"
	],
	"application/toolbook": [
		".tbk"
	],
	"application/vda": [
		".vda"
	],
	"application/vnd.3gpp.pic-bw-large": [
		".plb"
	],
	"application/vnd.3gpp.pic-bw-small": [
		".psb"
	],
	"application/vnd.3gpp.pic-bw-var": [
		".pvb"
	],
	"application/vnd.3gpp2.tcap": [
		".tcap"
	],
	"application/vnd.3m.post-it-notes": [
		".pwn"
	],
	"application/vnd.accpac.simply.aso": [
		".aso"
	],
	"application/vnd.accpac.simply.imp": [
		".imp"
	],
	"application/vnd.acucobol": [
		".acu"
	],
	"application/vnd.acucorp": [
		".atc"
	],
	"application/vnd.adobe.air-application-installer-package+zip": [
		".air"
	],
	"application/vnd.adobe.fxp": [
		".fxp"
	],
	"application/vnd.adobe.xdp+xml": [
		".xdp"
	],
	"application/vnd.adobe.xfdf": [
		".xfdf"
	],
	"application/vnd.ahead.space": [
		".ahead"
	],
	"application/vnd.airzip.filesecure.azf": [
		".azf"
	],
	"application/vnd.airzip.filesecure.azs": [
		".azs"
	],
	"application/vnd.amazon.ebook": [
		".azw"
	],
	"application/vnd.americandynamics.acc": [
		".acc"
	],
	"application/vnd.amiga.ami": [
		".ami"
	],
	"application/vnd.android.package-archive": [
		".apk"
	],
	"application/vnd.anser-web-certificate-issue-initiation": [
		".cii"
	],
	"application/vnd.anser-web-funds-transfer-initiation": [
		".fti"
	],
	"application/vnd.antix.game-component": [
		".atx"
	],
	"application/vnd.apple.installer+xml": [
		".mpkg"
	],
	"application/vnd.apple.mpegurl": [
		".m3u8"
	],
	"application/vnd.aristanetworks.swi": [
		".swi"
	],
	"application/vnd.audiograph": [
		".aep"
	],
	"application/vnd.blueice.multipass": [
		".mpm"
	],
	"application/vnd.bmi": [
		".bmi"
	],
	"application/vnd.businessobjects": [
		".rep"
	],
	"application/vnd.chemdraw+xml": [
		".cdxml"
	],
	"application/vnd.chipnuts.karaoke-mmd": [
		".mmd"
	],
	"application/vnd.cinderella": [
		".cdy"
	],
	"application/vnd.claymore": [
		".cla"
	],
	"application/vnd.cloanto.rp9": [
		".rp9"
	],
	"application/vnd.clonk.c4group": [
		".c4g"
	],
	"application/vnd.cluetrust.cartomobile-config": [
		".c11amc"
	],
	"application/vnd.cluetrust.cartomobile-config-pkg": [
		".c11amz"
	],
	"application/vnd.commonspace": [
		".csp"
	],
	"application/vnd.contact.cmsg": [
		".cdbcmsg"
	],
	"application/vnd.cosmocaller": [
		".cmc"
	],
	"application/vnd.crick.clicker": [
		".clkx"
	],
	"application/vnd.crick.clicker.keyboard": [
		".clkk"
	],
	"application/vnd.crick.clicker.palette": [
		".clkp"
	],
	"application/vnd.crick.clicker.template": [
		".clkt"
	],
	"application/vnd.crick.clicker.wordbank": [
		".clkw"
	],
	"application/vnd.criticaltools.wbs+xml": [
		".wbs"
	],
	"application/vnd.ctc-posml": [
		".pml"
	],
	"application/vnd.cups-ppd": [
		".ppd"
	],
	"application/vnd.curl.car": [
		".car"
	],
	"application/vnd.curl.pcurl": [
		".pcurl"
	],
	"application/vnd.data-vision.rdz": [
		".rdz"
	],
	"application/vnd.denovo.fcselayout-link": [
		".fe_launch"
	],
	"application/vnd.dna": [
		".dna"
	],
	"application/vnd.dolby.mlp": [
		".mlp"
	],
	"application/vnd.dpgraph": [
		".dpg"
	],
	"application/vnd.dreamfactory": [
		".dfac"
	],
	"application/vnd.dvb.ait": [
		".ait"
	],
	"application/vnd.dvb.service": [
		".svc"
	],
	"application/vnd.dynageo": [
		".geo"
	],
	"application/vnd.ecowin.chart": [
		".mag"
	],
	"application/vnd.enliven": [
		".nml"
	],
	"application/vnd.epson.esf": [
		".esf"
	],
	"application/vnd.epson.msf": [
		".msf"
	],
	"application/vnd.epson.quickanime": [
		".qam"
	],
	"application/vnd.epson.salt": [
		".slt"
	],
	"application/vnd.epson.ssf": [
		".ssf"
	],
	"application/vnd.eszigno3+xml": [
		".es3"
	],
	"application/vnd.ezpix-album": [
		".ez2"
	],
	"application/vnd.ezpix-package": [
		".ez3"
	],
	"application/vnd.fdf": [
		".fdf"
	],
	"application/vnd.fdsn.seed": [
		".seed"
	],
	"application/vnd.flographit": [
		".gph"
	],
	"application/vnd.fluxtime.clip": [
		".ftc"
	],
	"application/vnd.framemaker": [
		".fm"
	],
	"application/vnd.frogans.fnc": [
		".fnc"
	],
	"application/vnd.frogans.ltf": [
		".ltf"
	],
	"application/vnd.fsc.weblaunch": [
		".fsc"
	],
	"application/vnd.fujitsu.oasys": [
		".oas"
	],
	"application/vnd.fujitsu.oasys2": [
		".oa2"
	],
	"application/vnd.fujitsu.oasys3": [
		".oa3"
	],
	"application/vnd.fujitsu.oasysgp": [
		".fg5"
	],
	"application/vnd.fujitsu.oasysprs": [
		".bh2"
	],
	"application/vnd.fujixerox.ddd": [
		".ddd"
	],
	"application/vnd.fujixerox.docuworks": [
		".xdw"
	],
	"application/vnd.fujixerox.docuworks.binder": [
		".xbd"
	],
	"application/vnd.fuzzysheet": [
		".fzs"
	],
	"application/vnd.genomatix.tuxedo": [
		".txd"
	],
	"application/vnd.geogebra.file": [
		".ggb"
	],
	"application/vnd.geogebra.tool": [
		".ggt"
	],
	"application/vnd.geometry-explorer": [
		".gex"
	],
	"application/vnd.geonext": [
		".gxt"
	],
	"application/vnd.geoplan": [
		".g2w"
	],
	"application/vnd.geospace": [
		".g3w"
	],
	"application/vnd.gmx": [
		".gmx"
	],
	"application/vnd.google-earth.kml+xml": [
		".kml"
	],
	"application/vnd.google-earth.kmz": [
		".kmz"
	],
	"application/vnd.grafeq": [
		".gqf"
	],
	"application/vnd.groove-account": [
		".gac"
	],
	"application/vnd.groove-help": [
		".ghf"
	],
	"application/vnd.groove-identity-message": [
		".gim"
	],
	"application/vnd.groove-injector": [
		".grv"
	],
	"application/vnd.groove-tool-message": [
		".gtm"
	],
	"application/vnd.groove-tool-template": [
		".tpl"
	],
	"application/vnd.groove-vcard": [
		".vcg"
	],
	"application/vnd.hal+xml": [
		".hal"
	],
	"application/vnd.handheld-entertainment+xml": [
		".zmm"
	],
	"application/vnd.hbci": [
		".hbci"
	],
	"application/vnd.hhe.lesson-player": [
		".les"
	],
	"application/vnd.hp-hpgl": [
		".hpgl",
		".hgl",
		".hpg"
	],
	"application/vnd.hp-hpid": [
		".hpid"
	],
	"application/vnd.hp-hps": [
		".hps"
	],
	"application/vnd.hp-jlyt": [
		".jlt"
	],
	"application/vnd.hp-pcl": [
		".pcl"
	],
	"application/vnd.hp-pclxl": [
		".pclxl"
	],
	"application/vnd.hydrostatix.sof-data": [
		".sfd-hdstx"
	],
	"application/vnd.hzn-3d-crossword": [
		".x3d"
	],
	"application/vnd.ibm.minipay": [
		".mpy"
	],
	"application/vnd.ibm.modcap": [
		".afp"
	],
	"application/vnd.ibm.rights-management": [
		".irm"
	],
	"application/vnd.ibm.secure-container": [
		".sc"
	],
	"application/vnd.iccprofile": [
		".icc"
	],
	"application/vnd.igloader": [
		".igl"
	],
	"application/vnd.immervision-ivp": [
		".ivp"
	],
	"application/vnd.immervision-ivu": [
		".ivu"
	],
	"application/vnd.insors.igm": [
		".igm"
	],
	"application/vnd.intercon.formnet": [
		".xpw"
	],
	"application/vnd.intergeo": [
		".i2g"
	],
	"application/vnd.intu.qbo": [
		".qbo"
	],
	"application/vnd.intu.qfx": [
		".qfx"
	],
	"application/vnd.ipunplugged.rcprofile": [
		".rcprofile"
	],
	"application/vnd.irepository.package+xml": [
		".irp"
	],
	"application/vnd.is-xpr": [
		".xpr"
	],
	"application/vnd.isac.fcs": [
		".fcs"
	],
	"application/vnd.jam": [
		".jam"
	],
	"application/vnd.jcp.javame.midlet-rms": [
		".rms"
	],
	"application/vnd.jisp": [
		".jisp"
	],
	"application/vnd.joost.joda-archive": [
		".joda"
	],
	"application/vnd.kahootz": [
		".ktz"
	],
	"application/vnd.kde.karbon": [
		".karbon"
	],
	"application/vnd.kde.kchart": [
		".chrt"
	],
	"application/vnd.kde.kformula": [
		".kfo"
	],
	"application/vnd.kde.kivio": [
		".flw"
	],
	"application/vnd.kde.kontour": [
		".kon"
	],
	"application/vnd.kde.kpresenter": [
		".kpr"
	],
	"application/vnd.kde.kspread": [
		".ksp"
	],
	"application/vnd.kde.kword": [
		".kwd"
	],
	"application/vnd.kenameaapp": [
		".htke"
	],
	"application/vnd.kidspiration": [
		".kia"
	],
	"application/vnd.kinar": [
		".kne"
	],
	"application/vnd.koan": [
		".skp"
	],
	"application/vnd.kodak-descriptor": [
		".sse"
	],
	"application/vnd.las.las+xml": [
		".lasxml"
	],
	"application/vnd.llamagraphics.life-balance.desktop": [
		".lbd"
	],
	"application/vnd.llamagraphics.life-balance.exchange+xml": [
		".lbe"
	],
	"application/vnd.lotus-1-2-3": [
		".123"
	],
	"application/vnd.lotus-approach": [
		".apr"
	],
	"application/vnd.lotus-freelance": [
		".pre"
	],
	"application/vnd.lotus-notes": [
		".nsf"
	],
	"application/vnd.lotus-organizer": [
		".org"
	],
	"application/vnd.lotus-screencam": [
		".scm"
	],
	"application/vnd.lotus-wordpro": [
		".lwp"
	],
	"application/vnd.macports.portpkg": [
		".portpkg"
	],
	"application/vnd.mcd": [
		".mcd"
	],
	"application/vnd.medcalcdata": [
		".mc1"
	],
	"application/vnd.mediastation.cdkey": [
		".cdkey"
	],
	"application/vnd.mfer": [
		".mwf"
	],
	"application/vnd.mfmp": [
		".mfm"
	],
	"application/vnd.micrografx.flo": [
		".flo"
	],
	"application/vnd.micrografx.igx": [
		".igx"
	],
	"application/vnd.mif": [
		".mif"
	],
	"application/vnd.mobius.daf": [
		".daf"
	],
	"application/vnd.mobius.dis": [
		".dis"
	],
	"application/vnd.mobius.mbk": [
		".mbk"
	],
	"application/vnd.mobius.mqy": [
		".mqy"
	],
	"application/vnd.mobius.msl": [
		".msl"
	],
	"application/vnd.mobius.plc": [
		".plc"
	],
	"application/vnd.mobius.txf": [
		".txf"
	],
	"application/vnd.mophun.application": [
		".mpn"
	],
	"application/vnd.mophun.certificate": [
		".mpc"
	],
	"application/vnd.mozilla.xul+xml": [
		".xul"
	],
	"application/vnd.ms-artgalry": [
		".cil"
	],
	"application/vnd.ms-cab-compressed": [
		".cab"
	],
	"application/vnd.ms-excel": [
		".xls"
	],
	"application/vnd.ms-excel.addin.macroenabled.12": [
		".xlam"
	],
	"application/vnd.ms-excel.sheet.binary.macroenabled.12": [
		".xlsb"
	],
	"application/vnd.ms-excel.sheet.macroenabled.12": [
		".xlsm"
	],
	"application/vnd.ms-excel.template.macroenabled.12": [
		".xltm"
	],
	"application/vnd.ms-fontobject": [
		".eot"
	],
	"application/vnd.ms-htmlhelp": [
		".chm"
	],
	"application/vnd.ms-ims": [
		".ims"
	],
	"application/vnd.ms-lrm": [
		".lrm"
	],
	"application/vnd.ms-officetheme": [
		".thmx"
	],
	"application/vnd.ms-pki.certstore": [
		".sst"
	],
	"application/vnd.ms-pki.pko": [
		".pko"
	],
	"application/vnd.ms-pki.seccat": [
		".cat"
	],
	"application/vnd.ms-pki.stl": [
		".stl"
	],
	"application/vnd.ms-powerpoint": [
		".ppt",
		".pwz",
		".ppa"
	],
	"application/vnd.ms-powerpoint.addin.macroenabled.12": [
		".ppam"
	],
	"application/vnd.ms-powerpoint.presentation.macroenabled.12": [
		".pptm"
	],
	"application/vnd.ms-powerpoint.slide.macroenabled.12": [
		".sldm"
	],
	"application/vnd.ms-powerpoint.slideshow.macroenabled.12": [
		".ppsm"
	],
	"application/vnd.ms-powerpoint.template.macroenabled.12": [
		".potm"
	],
	"application/vnd.ms-project": [
		".mpp"
	],
	"application/vnd.ms-word.document.macroenabled.12": [
		".docm"
	],
	"application/vnd.ms-word.template.macroenabled.12": [
		".dotm"
	],
	"application/vnd.ms-works": [
		".wps"
	],
	"application/vnd.ms-wpl": [
		".wpl"
	],
	"application/vnd.ms-xpsdocument": [
		".xps"
	],
	"application/vnd.mseq": [
		".mseq"
	],
	"application/vnd.musician": [
		".mus"
	],
	"application/vnd.muvee.style": [
		".msty"
	],
	"application/vnd.neurolanguage.nlu": [
		".nlu"
	],
	"application/vnd.noblenet-directory": [
		".nnd"
	],
	"application/vnd.noblenet-sealer": [
		".nns"
	],
	"application/vnd.noblenet-web": [
		".nnw"
	],
	"application/vnd.nokia.configuration-message": [
		".ncm"
	],
	"application/vnd.nokia.n-gage.data": [
		".ngdat"
	],
	"application/vnd.nokia.n-gage.symbian.install": [
		".n-gage"
	],
	"application/vnd.nokia.radio-preset": [
		".rpst"
	],
	"application/vnd.nokia.radio-presets": [
		".rpss"
	],
	"application/vnd.novadigm.edm": [
		".edm"
	],
	"application/vnd.novadigm.edx": [
		".edx"
	],
	"application/vnd.novadigm.ext": [
		".ext"
	],
	"application/vnd.oasis.opendocument.chart": [
		".odc"
	],
	"application/vnd.oasis.opendocument.chart-template": [
		".otc"
	],
	"application/vnd.oasis.opendocument.database": [
		".odb"
	],
	"application/vnd.oasis.opendocument.formula": [
		".odf"
	],
	"application/vnd.oasis.opendocument.formula-template": [
		".odft"
	],
	"application/vnd.oasis.opendocument.graphics": [
		".odg"
	],
	"application/vnd.oasis.opendocument.graphics-flat-xml": [
		".fodg"
	],
	"application/vnd.oasis.opendocument.graphics-template": [
		".otg"
	],
	"application/vnd.oasis.opendocument.image": [
		".odi"
	],
	"application/vnd.oasis.opendocument.image-template": [
		".oti"
	],
	"application/vnd.oasis.opendocument.presentation": [
		".odp"
	],
	"application/vnd.oasis.opendocument.presentation-flat-xml": [
		".fodp"
	],
	"application/vnd.oasis.opendocument.presentation-template": [
		".otp"
	],
	"application/vnd.oasis.opendocument.spreadsheet": [
		".ods"
	],
	"application/vnd.oasis.opendocument.spreadsheet-flat-xml": [
		".fods"
	],
	"application/vnd.oasis.opendocument.spreadsheet-template": [
		".ots"
	],
	"application/vnd.oasis.opendocument.text": [
		".odt"
	],
	"application/vnd.oasis.opendocument.text-flat-xml": [
		".fodt"
	],
	"application/vnd.oasis.opendocument.text-master": [
		".odm"
	],
	"application/vnd.oasis.opendocument.text-template": [
		".ott"
	],
	"application/vnd.oasis.opendocument.text-web": [
		".oth"
	],
	"application/vnd.olpc-sugar": [
		".xo"
	],
	"application/vnd.oma.dd2+xml": [
		".dd2"
	],
	"application/vnd.openofficeorg.extension": [
		".oxt"
	],
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": [
		".pptx"
	],
	"application/vnd.openxmlformats-officedocument.presentationml.slide": [
		".sldx"
	],
	"application/vnd.openxmlformats-officedocument.presentationml.slideshow": [
		".ppsx"
	],
	"application/vnd.openxmlformats-officedocument.presentationml.template": [
		".potx"
	],
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [
		".xlsx"
	],
	"application/vnd.openxmlformats-officedocument.spreadsheetml.template": [
		".xltx"
	],
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": [
		".docx"
	],
	"application/vnd.openxmlformats-officedocument.wordprocessingml.template": [
		".dotx"
	],
	"application/vnd.osgeo.mapguide.package": [
		".mgp"
	],
	"application/vnd.osgi.dp": [
		".dp"
	],
	"application/vnd.palm": [
		".pdb"
	],
	"application/vnd.pawaafile": [
		".paw"
	],
	"application/vnd.pg.format": [
		".str"
	],
	"application/vnd.pg.osasli": [
		".ei6"
	],
	"application/vnd.picsel": [
		".efif"
	],
	"application/vnd.pmi.widget": [
		".wg"
	],
	"application/vnd.pocketlearn": [
		".plf"
	],
	"application/vnd.powerbuilder6": [
		".pbd"
	],
	"application/vnd.previewsystems.box": [
		".box"
	],
	"application/vnd.proteus.magazine": [
		".mgz"
	],
	"application/vnd.publishare-delta-tree": [
		".qps"
	],
	"application/vnd.pvi.ptid1": [
		".ptid"
	],
	"application/vnd.quark.quarkxpress": [
		".qxd"
	],
	"application/vnd.realvnc.bed": [
		".bed"
	],
	"application/vnd.recordare.musicxml": [
		".mxl"
	],
	"application/vnd.recordare.musicxml+xml": [
		".musicxml"
	],
	"application/vnd.rig.cryptonote": [
		".cryptonote"
	],
	"application/vnd.rim.cod": [
		".cod"
	],
	"application/vnd.rn-realmedia": [
		".rm"
	],
	"application/vnd.rn-realplayer": [
		".rnx"
	],
	"application/vnd.route66.link66+xml": [
		".link66"
	],
	"application/vnd.sailingtracker.track": [
		".st"
	],
	"application/vnd.seemail": [
		".see"
	],
	"application/vnd.sema": [
		".sema"
	],
	"application/vnd.semd": [
		".semd"
	],
	"application/vnd.semf": [
		".semf"
	],
	"application/vnd.shana.informed.formdata": [
		".ifm"
	],
	"application/vnd.shana.informed.formtemplate": [
		".itp"
	],
	"application/vnd.shana.informed.interchange": [
		".iif"
	],
	"application/vnd.shana.informed.package": [
		".ipk"
	],
	"application/vnd.simtech-mindmapper": [
		".twd"
	],
	"application/vnd.smaf": [
		".mmf"
	],
	"application/vnd.smart.teacher": [
		".teacher"
	],
	"application/vnd.solent.sdkm+xml": [
		".sdkm"
	],
	"application/vnd.spotfire.dxp": [
		".dxp"
	],
	"application/vnd.spotfire.sfs": [
		".sfs"
	],
	"application/vnd.stardivision.calc": [
		".sdc"
	],
	"application/vnd.stardivision.chart": [
		".sds"
	],
	"application/vnd.stardivision.draw": [
		".sda"
	],
	"application/vnd.stardivision.impress": [
		".sdd"
	],
	"application/vnd.stardivision.math": [
		".smf"
	],
	"application/vnd.stardivision.writer": [
		".sdw",
		".vor"
	],
	"application/vnd.stardivision.writer-global": [
		".sgl"
	],
	"application/vnd.stepmania.stepchart": [
		".sm"
	],
	"application/vnd.sun.xml.calc": [
		".sxc"
	],
	"application/vnd.sun.xml.calc.template": [
		".stc"
	],
	"application/vnd.sun.xml.draw": [
		".sxd"
	],
	"application/vnd.sun.xml.draw.template": [
		".std"
	],
	"application/vnd.sun.xml.impress": [
		".sxi"
	],
	"application/vnd.sun.xml.impress.template": [
		".sti"
	],
	"application/vnd.sun.xml.math": [
		".sxm"
	],
	"application/vnd.sun.xml.writer": [
		".sxw"
	],
	"application/vnd.sun.xml.writer.global": [
		".sxg"
	],
	"application/vnd.sun.xml.writer.template": [
		".stw"
	],
	"application/vnd.sus-calendar": [
		".sus"
	],
	"application/vnd.svd": [
		".svd"
	],
	"application/vnd.symbian.install": [
		".sis"
	],
	"application/vnd.syncml+xml": [
		".xsm"
	],
	"application/vnd.syncml.dm+wbxml": [
		".bdm"
	],
	"application/vnd.syncml.dm+xml": [
		".xdm"
	],
	"application/vnd.tao.intent-module-archive": [
		".tao"
	],
	"application/vnd.tmobile-livetv": [
		".tmo"
	],
	"application/vnd.trid.tpt": [
		".tpt"
	],
	"application/vnd.triscape.mxs": [
		".mxs"
	],
	"application/vnd.trueapp": [
		".tra"
	],
	"application/vnd.ufdl": [
		".ufd"
	],
	"application/vnd.uiq.theme": [
		".utz"
	],
	"application/vnd.umajin": [
		".umj"
	],
	"application/vnd.unity": [
		".unityweb"
	],
	"application/vnd.uoml+xml": [
		".uoml"
	],
	"application/vnd.vcx": [
		".vcx"
	],
	"application/vnd.visio": [
		".vsd"
	],
	"application/vnd.visionary": [
		".vis"
	],
	"application/vnd.vsf": [
		".vsf"
	],
	"application/vnd.wap.wbxml": [
		".wbxml"
	],
	"application/vnd.wap.wmlc": [
		".wmlc"
	],
	"application/vnd.wap.wmlscriptc": [
		".wmlsc"
	],
	"application/vnd.webturbo": [
		".wtb"
	],
	"application/vnd.wolfram.player": [
		".nbp"
	],
	"application/vnd.wordperfect": [
		".wpd"
	],
	"application/vnd.wqd": [
		".wqd"
	],
	"application/vnd.wt.stf": [
		".stf"
	],
	"application/vnd.xara": [
		".web",
		".xar"
	],
	"application/vnd.xfdl": [
		".xfdl"
	],
	"application/vnd.yamaha.hv-dic": [
		".hvd"
	],
	"application/vnd.yamaha.hv-script": [
		".hvs"
	],
	"application/vnd.yamaha.hv-voice": [
		".hvp"
	],
	"application/vnd.yamaha.openscoreformat": [
		".osf"
	],
	"application/vnd.yamaha.openscoreformat.osfpvg+xml": [
		".osfpvg"
	],
	"application/vnd.yamaha.smaf-audio": [
		".saf"
	],
	"application/vnd.yamaha.smaf-phrase": [
		".spf"
	],
	"application/vnd.yellowriver-custom-menu": [
		".cmp"
	],
	"application/vnd.zul": [
		".zir"
	],
	"application/vnd.zzazz.deck+xml": [
		".zaz"
	],
	"application/vocaltec-media-desc": [
		".vmd"
	],
	"application/vocaltec-media-file": [
		".vmf"
	],
	"application/voicexml+xml": [
		".vxml"
	],
	"application/widget": [
		".wgt"
	],
	"application/winhlp": [
		".hlp"
	],
	"application/wordperfect": [
		".wp5",
		".wp6",
		".wp"
	],
	"application/wordperfect6.0": [
		".w60"
	],
	"application/wordperfect6.1": [
		".w61"
	],
	"application/wsdl+xml": [
		".wsdl"
	],
	"application/wspolicy+xml": [
		".wspolicy"
	],
	"application/x-123": [
		".wk",
		".wk1"
	],
	"application/x-7z-compressed": [
		".7z"
	],
	"application/x-abiword": [
		".abw"
	],
	"application/x-ace-compressed": [
		".ace"
	],
	"application/x-aim": [
		".aim"
	],
	"application/x-apple-diskimage": [
		".dmg"
	],
	"application/x-authorware-bin": [
		".aab"
	],
	"application/x-authorware-map": [
		".aam"
	],
	"application/x-authorware-seg": [
		".aas"
	],
	"application/x-bcpio": [
		".bcpio"
	],
	"application/x-bittorrent": [
		".torrent"
	],
	"application/x-bsh": [
		".bsh"
	],
	"application/x-bytecode.elisp (compiled elisp)": [
		".elc"
	],
	"application/x-bzip": [
		".bz"
	],
	"application/x-bzip2": [
		".bz2",
		".boz"
	],
	"application/x-cbr": [
		".cbr"
	],
	"application/x-cbz": [
		".cbz"
	],
	"application/x-cdf": [
		".cda"
	],
	"application/x-cdlink": [
		".vcd"
	],
	"application/x-chat": [
		".chat",
		".cha"
	],
	"application/x-chess-pgn": [
		".pgn"
	],
	"application/x-cocoa": [
		".cco"
	],
	"application/x-compress": [
		".z"
	],
	"application/x-compressed": [
		".gz"
	],
	"application/x-comsol": [
		".mph"
	],
	"application/x-conference": [
		".nsc"
	],
	"application/x-cpio": [
		".cpio"
	],
	"application/x-csh": [
		".csh"
	],
	"application/x-debian-package": [
		".deb",
		".udeb"
	],
	"application/x-deepv": [
		".deepv"
	],
	"application/x-director": [
		".dxr",
		".dir",
		".dcr"
	],
	"application/x-dms": [
		".dms"
	],
	"application/x-doom": [
		".wad"
	],
	"application/x-dtbncx+xml": [
		".ncx"
	],
	"application/x-dtbook+xml": [
		".dtb"
	],
	"application/x-dtbresource+xml": [
		".res"
	],
	"application/x-dvi": [
		".dvi"
	],
	"application/x-envoy": [
		".env"
	],
	"application/x-font": [
		".pfb",
		".pcf.z"
	],
	"application/x-font-bdf": [
		".bdf"
	],
	"application/x-font-ghostscript": [
		".gsf"
	],
	"application/x-font-linux-psf": [
		".psf"
	],
	"application/x-font-otf": [
		".otf"
	],
	"application/x-font-pcf": [
		".pcf"
	],
	"application/x-font-snf": [
		".snf"
	],
	"application/x-font-ttf": [
		".ttf"
	],
	"application/x-font-type1": [
		".pfa"
	],
	"application/x-font-woff": [
		".woff"
	],
	"application/x-futuresplash": [
		".spl"
	],
	"application/x-ganttproject": [
		".gan"
	],
	"application/x-gnumeric": [
		".gnumeric"
	],
	"application/x-go-sgf": [
		".sgf"
	],
	"application/x-graphing-calculator": [
		".gcf"
	],
	"application/x-gsp": [
		".gsp"
	],
	"application/x-gss": [
		".gss"
	],
	"application/x-gtar": [
		".gtar"
	],
	"application/x-gtar-compressed": [
		".taz"
	],
	"application/x-gzip": [
		".gzip"
	],
	"application/x-hdf": [
		".hdf"
	],
	"application/x-helpfile": [
		".help"
	],
	"application/x-httpd-eruby": [
		".rhtml"
	],
	"application/x-httpd-imap": [
		".imap"
	],
	"application/x-httpd-php": [
		".php",
		".phtml",
		".pht"
	],
	"application/x-httpd-php-source": [
		".phps"
	],
	"application/x-httpd-php3": [
		".php3"
	],
	"application/x-httpd-php3-preprocessed": [
		".php3p"
	],
	"application/x-httpd-php4": [
		".php4"
	],
	"application/x-httpd-php5": [
		".php5"
	],
	"application/x-ica": [
		".ica"
	],
	"application/x-ima": [
		".ima"
	],
	"application/x-info": [
		".info"
	],
	"application/x-internet-signup": [
		".isp"
	],
	"application/x-internett-signup": [
		".ins"
	],
	"application/x-inventor": [
		".iv"
	],
	"application/x-ip2": [
		".ip"
	],
	"application/x-iphone": [
		".iii"
	],
	"application/x-iso9660-image": [
		".iso"
	],
	"application/x-java-commerce": [
		".jcm"
	],
	"application/x-java-jnlp-file": [
		".jnlp"
	],
	"application/x-jmol": [
		".jmz"
	],
	"application/x-killustrator": [
		".kil"
	],
	"application/x-koan": [
		".skd",
		".skt",
		".skm"
	],
	"application/x-kpresenter": [
		".kpt"
	],
	"application/x-ksh": [
		".ksh"
	],
	"application/x-kword": [
		".kwt"
	],
	"application/x-latex": [
		".latex",
		".ltx"
	],
	"application/x-lisp": [
		".lsp"
	],
	"application/x-livescreen": [
		".ivy"
	],
	"application/x-lotus": [
		".wq1"
	],
	"application/x-lyx": [
		".lyx"
	],
	"application/x-magic-cap-package-1.0": [
		".mc$"
	],
	"application/x-maker": [
		".maker",
		".fbdoc",
		".frame",
		".fb",
		".frm"
	],
	"application/x-midi": [
		".midi"
	],
	"application/x-mix-transfer": [
		".nix"
	],
	"application/x-mobipocket-ebook": [
		".prc"
	],
	"application/x-mplayer2": [
		".asx"
	],
	"application/x-ms-application": [
		".application"
	],
	"application/x-ms-wmd": [
		".wmd"
	],
	"application/x-ms-wmz": [
		".wmz"
	],
	"application/x-ms-xbap": [
		".xbap"
	],
	"application/x-msaccess": [
		".mdb"
	],
	"application/x-msbinder": [
		".obd"
	],
	"application/x-mscardfile": [
		".crd"
	],
	"application/x-msclip": [
		".clp"
	],
	"application/x-msdos-program": [
		".bat",
		".dll"
	],
	"application/x-msdownload": [
		".exe"
	],
	"application/x-msi": [
		".msi"
	],
	"application/x-msmediaview": [
		".mvb"
	],
	"application/x-msmetafile": [
		".wmf"
	],
	"application/x-msmoney": [
		".mny"
	],
	"application/x-mspublisher": [
		".pub"
	],
	"application/x-msschedule": [
		".scd"
	],
	"application/x-msterminal": [
		".trm"
	],
	"application/x-mswrite": [
		".wri"
	],
	"application/x-navi-animation": [
		".ani"
	],
	"application/x-navidoc": [
		".nvd"
	],
	"application/x-navimap": [
		".map"
	],
	"application/x-netcdf": [
		".nc"
	],
	"application/x-newton-compatible-pkg": [
		".pkg"
	],
	"application/x-nokia-9000-communicator-add-on-software": [
		".aos"
	],
	"application/x-ns-proxy-autoconfig": [
		".pac",
		".dat"
	],
	"application/x-nwc": [
		".nwc"
	],
	"application/x-omc": [
		".omc"
	],
	"application/x-omcdatamaker": [
		".omcd"
	],
	"application/x-omcregerator": [
		".omcr"
	],
	"application/x-oz-application": [
		".oza"
	],
	"application/x-pagemaker": [
		".pm5",
		".pm4"
	],
	"application/x-pixclscript": [
		".plx"
	],
	"application/x-pkcs12": [
		".p12"
	],
	"application/x-pkcs7-certificates": [
		".p7b",
		".spc"
	],
	"application/x-pkcs7-certreqresp": [
		".p7r"
	],
	"application/x-pkcs7-signature": [
		".p7a"
	],
	"application/x-project": [
		".mpx",
		".mpv",
		".mpt"
	],
	"application/x-python-code": [
		".pyo"
	],
	"application/x-qgis": [
		".shp",
		".shx",
		".qgs"
	],
	"application/x-qpro": [
		".wb1"
	],
	"application/x-quicktimeplayer": [
		".qtl"
	],
	"application/x-rar-compressed": [
		".rar"
	],
	"application/x-rdp": [
		".rdp"
	],
	"application/x-ruby": [
		".rb"
	],
	"application/x-scilab": [
		".sci",
		".sce"
	],
	"application/x-seelogo": [
		".sl"
	],
	"application/x-sh": [
		".sh"
	],
	"application/x-shar": [
		".shar"
	],
	"application/x-shockwave-flash": [
		".swf",
		".swfl"
	],
	"application/x-silverlight": [
		".scr"
	],
	"application/x-silverlight-app": [
		".xap"
	],
	"application/x-sprite": [
		".spr",
		".sprite"
	],
	"application/x-sql": [
		".sql"
	],
	"application/x-stuffit": [
		".sit"
	],
	"application/x-stuffitx": [
		".sitx"
	],
	"application/x-sv4cpio": [
		".sv4cpio"
	],
	"application/x-sv4crc": [
		".sv4crc"
	],
	"application/x-tar": [
		".tar"
	],
	"application/x-tbook": [
		".sbk"
	],
	"application/x-tcl": [
		".tcl"
	],
	"application/x-tex": [
		".tex"
	],
	"application/x-tex-gf": [
		".gf"
	],
	"application/x-tex-pk": [
		".pk"
	],
	"application/x-tex-tfm": [
		".tfm"
	],
	"application/x-texinfo": [
		".texi",
		".texinfo"
	],
	"application/x-trash": [
		".bak",
		".old",
		".%",
		".~",
		".sik"
	],
	"application/x-troff": [
		".roff",
		".tr"
	],
	"application/x-troff-man": [
		".man"
	],
	"application/x-troff-me": [
		".me"
	],
	"application/x-troff-ms": [
		".ms"
	],
	"application/x-ustar": [
		".ustar"
	],
	"application/x-visio": [
		".vsw",
		".vst"
	],
	"application/x-vnd.audioexplosion.mzz": [
		".mzz"
	],
	"application/x-vnd.ls-xpix": [
		".xpix"
	],
	"application/x-vrml": [
		".vrml"
	],
	"application/x-wais-source": [
		".wsrc",
		".src"
	],
	"application/x-wingz": [
		".wz"
	],
	"application/x-wintalk": [
		".wtk"
	],
	"application/x-world": [
		".svr"
	],
	"application/x-x509-ca-cert": [
		".der"
	],
	"application/x-xcf": [
		".xcf"
	],
	"application/x-xfig": [
		".fig"
	],
	"application/x-xpinstall": [
		".xpi"
	],
	"application/xcap-diff+xml": [
		".xdf"
	],
	"application/xenc+xml": [
		".xenc"
	],
	"application/xhtml+xml": [
		".xhtml",
		".xht"
	],
	"application/xml": [
		".xml",
		".xsd",
		".xsl"
	],
	"application/xml-dtd": [
		".dtd"
	],
	"application/xop+xml": [
		".xop"
	],
	"application/xslt+xml": [
		".xslt"
	],
	"application/xspf+xml": [
		".xspf"
	],
	"application/xv+xml": [
		".mxml"
	],
	"application/yang": [
		".yang"
	],
	"application/yin+xml": [
		".yin"
	],
	"application/zip": [
		".zip"
	],
	"audio/adpcm": [
		".adp"
	],
	"audio/aiff": [
		".aifc",
		".aiff"
	],
	"audio/amr": [
		".amr"
	],
	"audio/amr-wb": [
		".awb"
	],
	"audio/annodex": [
		".axa"
	],
	"audio/basic": [
		".snd",
		".au"
	],
	"audio/csound": [
		".orc",
		".csd",
		".sco"
	],
	"audio/flac": [
		".flac"
	],
	"audio/it": [
		".it"
	],
	"audio/make": [
		".funk",
		".my",
		".pfunk"
	],
	"audio/mid": [
		".rmi"
	],
	"audio/midi": [
		".kar",
		".mid"
	],
	"audio/mod": [
		".mod"
	],
	"audio/mp4": [
		".mp4a"
	],
	"audio/mpeg": [
		".mp3",
		".mpg",
		".mp2",
		".m4a",
		".mpga",
		".mpa",
		".mpega",
		".m2a"
	],
	"audio/nspaudio": [
		".lma",
		".la"
	],
	"audio/ogg": [
		".ogg",
		".oga",
		".spx"
	],
	"audio/s3m": [
		".s3m"
	],
	"audio/tsp-audio": [
		".tsi"
	],
	"audio/vnd.dece.audio": [
		".uva"
	],
	"audio/vnd.digital-winds": [
		".eol"
	],
	"audio/vnd.dra": [
		".dra"
	],
	"audio/vnd.dts": [
		".dts"
	],
	"audio/vnd.dts.hd": [
		".dtshd"
	],
	"audio/vnd.lucent.voice": [
		".lvp"
	],
	"audio/vnd.ms-playready.media.pya": [
		".pya"
	],
	"audio/vnd.nuera.ecelp4800": [
		".ecelp4800"
	],
	"audio/vnd.nuera.ecelp7470": [
		".ecelp7470"
	],
	"audio/vnd.nuera.ecelp9600": [
		".ecelp9600"
	],
	"audio/vnd.qcelp": [
		".qcp"
	],
	"audio/vnd.rip": [
		".rip"
	],
	"audio/voc": [
		".voc"
	],
	"audio/voxware": [
		".vox"
	],
	"audio/webm": [
		".weba"
	],
	"audio/x-aac": [
		".aac"
	],
	"audio/x-aiff": [
		".aif"
	],
	"audio/x-gsm": [
		".gsm",
		".gsd"
	],
	"audio/x-liveaudio": [
		".lam"
	],
	"audio/x-mpegurl": [
		".m3u"
	],
	"audio/x-ms-wax": [
		".wax"
	],
	"audio/x-ms-wma": [
		".wma"
	],
	"audio/x-pn-realaudio": [
		".ram",
		".ra",
		".rmm"
	],
	"audio/x-pn-realaudio-plugin": [
		".rmp",
		".rpm"
	],
	"audio/x-psid": [
		".sid"
	],
	"audio/x-sd2": [
		".sd2"
	],
	"audio/x-twinvq": [
		".vqf"
	],
	"audio/x-twinvq-plugin": [
		".vqe",
		".vql"
	],
	"audio/x-vnd.audioexplosion.mjuicemediafile": [
		".mjf"
	],
	"audio/x-wav": [
		".wav"
	],
	"audio/xm": [
		".xm"
	],
	"chemical/x-alchemy": [
		".alc"
	],
	"chemical/x-cache": [
		".cache",
		".cac"
	],
	"chemical/x-cache-csf": [
		".csf"
	],
	"chemical/x-cactvs-binary": [
		".cbin",
		".ctab",
		".cascii"
	],
	"chemical/x-cdx": [
		".cdx"
	],
	"chemical/x-chem3d": [
		".c3d"
	],
	"chemical/x-cif": [
		".cif"
	],
	"chemical/x-cmdf": [
		".cmdf"
	],
	"chemical/x-cml": [
		".cml"
	],
	"chemical/x-compass": [
		".cpa"
	],
	"chemical/x-crossfire": [
		".bsd"
	],
	"chemical/x-csml": [
		".csml",
		".csm"
	],
	"chemical/x-ctx": [
		".ctx"
	],
	"chemical/x-cxf": [
		".cef",
		".cxf"
	],
	"chemical/x-embl-dl-nucleotide": [
		".emb",
		".embl"
	],
	"chemical/x-gamess-input": [
		".inp",
		".gamin",
		".gam"
	],
	"chemical/x-gaussian-checkpoint": [
		".fch",
		".fchk"
	],
	"chemical/x-gaussian-cube": [
		".cub"
	],
	"chemical/x-gaussian-input": [
		".gjc",
		".gau",
		".gjf"
	],
	"chemical/x-gaussian-log": [
		".gal"
	],
	"chemical/x-gcg8-sequence": [
		".gcg"
	],
	"chemical/x-genbank": [
		".gen"
	],
	"chemical/x-hin": [
		".hin"
	],
	"chemical/x-isostar": [
		".ist",
		".istr"
	],
	"chemical/x-jcamp-dx": [
		".dx",
		".jdx"
	],
	"chemical/x-kinemage": [
		".kin"
	],
	"chemical/x-macmolecule": [
		".mcm"
	],
	"chemical/x-macromodel-input": [
		".mmod"
	],
	"chemical/x-mdl-molfile": [
		".mol"
	],
	"chemical/x-mdl-rdfile": [
		".rd"
	],
	"chemical/x-mdl-rxnfile": [
		".rxn"
	],
	"chemical/x-mdl-sdfile": [
		".sdf",
		".sd"
	],
	"chemical/x-mdl-tgf": [
		".tgf"
	],
	"chemical/x-mmcif": [
		".mcif"
	],
	"chemical/x-mol2": [
		".mol2"
	],
	"chemical/x-molconn-Z": [
		".b"
	],
	"chemical/x-mopac-graph": [
		".gpt"
	],
	"chemical/x-mopac-input": [
		".mopcrt",
		".zmt",
		".mop"
	],
	"chemical/x-mopac-out": [
		".moo"
	],
	"chemical/x-ncbi-asn1": [
		".asn"
	],
	"chemical/x-ncbi-asn1-ascii": [
		".ent"
	],
	"chemical/x-ncbi-asn1-binary": [
		".val"
	],
	"chemical/x-rosdal": [
		".ros"
	],
	"chemical/x-swissprot": [
		".sw"
	],
	"chemical/x-vamas-iso14976": [
		".vms"
	],
	"chemical/x-xtel": [
		".xtel"
	],
	"chemical/x-xyz": [
		".xyz"
	],
	"i-world/i-vrml": [
		".ivr"
	],
	"image/bmp": [
		".bmp",
		".bm"
	],
	"image/cgm": [
		".cgm"
	],
	"image/cmu-raster": [
		".rast"
	],
	"image/florian": [
		".turbot"
	],
	"image/g3fax": [
		".g3"
	],
	"image/gif": [
		".gif"
	],
	"image/ief": [
		".iefs",
		".ief"
	],
	"image/jpeg": [
		".jpg",
		".jfif-tbnl",
		".jpeg",
		".jpe",
		".jfif"
	],
	"image/jutvision": [
		".jut"
	],
	"image/ktx": [
		".ktx"
	],
	"image/naplps": [
		".naplps",
		".nap"
	],
	"image/pict": [
		".pict"
	],
	"image/png": [
		".png",
		".x-png"
	],
	"image/prs.btif": [
		".btif"
	],
	"image/svg+xml": [
		".svgz",
		".svg"
	],
	"image/tiff": [
		".tif",
		".tiff"
	],
	"image/vasa": [
		".mcf"
	],
	"image/vnd.adobe.photoshop": [
		".psd"
	],
	"image/vnd.dece.graphic": [
		".uvi"
	],
	"image/vnd.djvu": [
		".djvu",
		".djv"
	],
	"image/vnd.dvb.subtitle": [
		".sub"
	],
	"image/vnd.dwg": [
		".dwg",
		".svf"
	],
	"image/vnd.dxf": [
		".dxf"
	],
	"image/vnd.fastbidsheet": [
		".fbs"
	],
	"image/vnd.fpx": [
		".fpx"
	],
	"image/vnd.fst": [
		".fst"
	],
	"image/vnd.fujixerox.edmics-mmr": [
		".mmr"
	],
	"image/vnd.fujixerox.edmics-rlc": [
		".rlc"
	],
	"image/vnd.ms-modi": [
		".mdi"
	],
	"image/vnd.net-fpx": [
		".npx"
	],
	"image/vnd.rn-realflash": [
		".rf"
	],
	"image/vnd.rn-realpix": [
		".rp"
	],
	"image/vnd.wap.wbmp": [
		".wbmp"
	],
	"image/vnd.xiff": [
		".xif"
	],
	"image/webp": [
		".webp"
	],
	"image/x-canon-cr2": [
		".cr2"
	],
	"image/x-canon-crw": [
		".crw"
	],
	"image/x-cmu-raster": [
		".ras"
	],
	"image/x-cmx": [
		".cmx"
	],
	"image/x-coreldraw": [
		".cdr"
	],
	"image/x-coreldrawpattern": [
		".pat"
	],
	"image/x-coreldrawtemplate": [
		".cdt"
	],
	"image/x-epson-erf": [
		".erf"
	],
	"image/x-freehand": [
		".fh"
	],
	"image/x-icon": [
		".ico"
	],
	"image/x-jg": [
		".art"
	],
	"image/x-jng": [
		".jng"
	],
	"image/x-jps": [
		".jps"
	],
	"image/x-niff": [
		".niff",
		".nif"
	],
	"image/x-nikon-nef": [
		".nef"
	],
	"image/x-olympus-orf": [
		".orf"
	],
	"image/x-pcx": [
		".pcx"
	],
	"image/x-pict": [
		".pic",
		".pct"
	],
	"image/x-portable-anymap": [
		".pnm"
	],
	"image/x-portable-bitmap": [
		".pbm"
	],
	"image/x-portable-graymap": [
		".pgm"
	],
	"image/x-portable-pixmap": [
		".ppm"
	],
	"image/x-quicktime": [
		".qti",
		".qtif",
		".qif"
	],
	"image/x-rgb": [
		".rgb"
	],
	"image/x-xbitmap": [
		".xbm"
	],
	"image/x-xpixmap": [
		".pm",
		".xpm"
	],
	"image/x-xwindowdump": [
		".xwd"
	],
	"message/rfc822": [
		".mht",
		".eml",
		".mime",
		".mhtml"
	],
	"model/iges": [
		".igs"
	],
	"model/mesh": [
		".mesh",
		".silo",
		".msh"
	],
	"model/vnd.collada+xml": [
		".dae"
	],
	"model/vnd.dwf": [
		".dwf"
	],
	"model/vnd.gdl": [
		".gdl"
	],
	"model/vnd.gtw": [
		".gtw"
	],
	"model/vnd.mts": [
		".mts"
	],
	"model/vnd.vtu": [
		".vtu"
	],
	"model/vrml": [
		".wrz",
		".wrl"
	],
	"model/x-pov": [
		".pov"
	],
	"model/x3d+binary": [
		".x3db"
	],
	"model/x3d+vrml": [
		".x3dv"
	],
	"paleovu/x-pv": [
		".pvu"
	],
	"text/asp": [
		".asp"
	],
	"text/cache-manifest": [
		".manifest"
	],
	"text/calendar": [
		".ics",
		".icz"
	],
	"text/css": [
		".css"
	],
	"text/csv": [
		".csv"
	],
	"text/h323": [
		".323"
	],
	"text/html": [
		".html",
		".acgi",
		".htm",
		".htmls",
		".htx",
		".shtml"
	],
	"text/iuls": [
		".uls"
	],
	"text/mathml": [
		".mml"
	],
	"text/n3": [
		".n3"
	],
	"text/pascal": [
		".pas"
	],
	"text/plain": [
		".txt",
		".m",
		".brf",
		".cc",
		".lst",
		".conf",
		".cxx",
		".asc",
		".jav",
		".mar",
		".list",
		".pl",
		".c++",
		".log",
		".def",
		".hh",
		".sdml",
		".h",
		".g",
		".f90",
		".idc",
		".for"
	],
	"text/plain-bas": [
		".par"
	],
	"text/prs.lines.tag": [
		".dsc"
	],
	"text/richtext": [
		".rtx",
		".rt"
	],
	"text/scriplet": [
		".wsc"
	],
	"text/scriptlet": [
		".sct"
	],
	"text/sgml": [
		".sgml",
		".sgm"
	],
	"text/tab-separated-values": [
		".tsv"
	],
	"text/texmacs": [
		".tm"
	],
	"text/troff": [
		".t"
	],
	"text/turtle": [
		".ttl"
	],
	"text/uri-list": [
		".uri",
		".unis",
		".uni",
		".uris"
	],
	"text/vnd.abc": [
		".abc"
	],
	"text/vnd.curl": [
		".curl"
	],
	"text/vnd.curl.dcurl": [
		".dcurl"
	],
	"text/vnd.curl.mcurl": [
		".mcurl"
	],
	"text/vnd.curl.scurl": [
		".scurl"
	],
	"text/vnd.fly": [
		".fly"
	],
	"text/vnd.fmi.flexstor": [
		".flx"
	],
	"text/vnd.graphviz": [
		".gv"
	],
	"text/vnd.in3d.3dml": [
		".3dml"
	],
	"text/vnd.in3d.spot": [
		".spot"
	],
	"text/vnd.sun.j2me.app-descriptor": [
		".jad"
	],
	"text/vnd.wap.wml": [
		".wml"
	],
	"text/vnd.wap.wmlscript": [
		".wmls"
	],
	"text/webviewhtml": [
		".htt"
	],
	"text/x-asm": [
		".asm",
		".s"
	],
	"text/x-audiosoft-intra": [
		".aip"
	],
	"text/x-bibtex": [
		".bib"
	],
	"text/x-c": [
		".cpp",
		".c"
	],
	"text/x-c++hdr": [
		".hpp",
		".hxx",
		".h++"
	],
	"text/x-component": [
		".htc"
	],
	"text/x-diff": [
		".diff",
		".patch"
	],
	"text/x-dsrc": [
		".d"
	],
	"text/x-fortran": [
		".f77",
		".f"
	],
	"text/x-haskell": [
		".hs"
	],
	"text/x-java-source,java": [
		".java"
	],
	"text/x-la-asf": [
		".lsx"
	],
	"text/x-literate-haskell": [
		".lhs"
	],
	"text/x-moc": [
		".moc"
	],
	"text/x-pascal": [
		".p"
	],
	"text/x-pcs-gcd": [
		".gcd"
	],
	"text/x-python": [
		".py"
	],
	"text/x-scala": [
		".scala"
	],
	"text/x-script": [
		".hlb"
	],
	"text/x-script.elisp": [
		".el"
	],
	"text/x-script.rexx": [
		".rexx"
	],
	"text/x-script.tcsh": [
		".tcsh"
	],
	"text/x-script.zsh": [
		".zsh"
	],
	"text/x-server-parsed-html": [
		".ssi"
	],
	"text/x-setext": [
		".etx"
	],
	"text/x-sfv": [
		".sfv"
	],
	"text/x-speech": [
		".talk"
	],
	"text/x-tcl": [
		".tk"
	],
	"text/x-tex": [
		".cls",
		".sty"
	],
	"text/x-uil": [
		".uil"
	],
	"text/x-uuencode": [
		".uu",
		".uue"
	],
	"text/x-vcalendar": [
		".vcs"
	],
	"text/x-vcard": [
		".vcf"
	],
	"text/yaml": [
		".yaml"
	],
	"video/3gpp": [
		".3gp"
	],
	"video/3gpp2": [
		".3g2"
	],
	"video/MP2T": [
		".ts"
	],
	"video/animaflex": [
		".afl"
	],
	"video/annodex": [
		".axv"
	],
	"video/avs-video": [
		".avs"
	],
	"video/dl": [
		".dl"
	],
	"video/gl": [
		".gl"
	],
	"video/h261": [
		".h261"
	],
	"video/h263": [
		".h263"
	],
	"video/h264": [
		".h264"
	],
	"video/jpeg": [
		".jpgv"
	],
	"video/jpm": [
		".jpm"
	],
	"video/mj2": [
		".mj2"
	],
	"video/mp4": [
		".mp4"
	],
	"video/mpeg": [
		".mpeg",
		".m2v",
		".m1v",
		".mpe"
	],
	"video/ogg": [
		".ogv"
	],
	"video/quicktime": [
		".mov",
		".moov",
		".qt"
	],
	"video/vdo": [
		".vdo"
	],
	"video/vivo": [
		".vivo"
	],
	"video/vnd.dece.hd": [
		".uvh"
	],
	"video/vnd.dece.mobile": [
		".uvm"
	],
	"video/vnd.dece.pd": [
		".uvp"
	],
	"video/vnd.dece.sd": [
		".uvs"
	],
	"video/vnd.dece.video": [
		".uvv"
	],
	"video/vnd.fvt": [
		".fvt"
	],
	"video/vnd.mpegurl": [
		".mxu"
	],
	"video/vnd.ms-playready.media.pyv": [
		".pyv"
	],
	"video/vnd.rn-realvideo": [
		".rv"
	],
	"video/vnd.uvvu.mp4": [
		".uvu"
	],
	"video/vnd.vivo": [
		".viv"
	],
	"video/vosaic": [
		".vos"
	],
	"video/webm": [
		".webm"
	],
	"video/x-amt-demorun": [
		".xdr"
	],
	"video/x-amt-showrun": [
		".xsr"
	],
	"video/x-atomic3d-feature": [
		".fmf"
	],
	"video/x-dv": [
		".dif",
		".dv"
	],
	"video/x-f4v": [
		".f4v"
	],
	"video/x-fli": [
		".fli"
	],
	"video/x-flv": [
		".flv"
	],
	"video/x-isvideo": [
		".isu"
	],
	"video/x-la-asf": [
		".lsf"
	],
	"video/x-m4v": [
		".m4v"
	],
	"video/x-matroska": [
		".mkv"
	],
	"video/x-mng": [
		".mng"
	],
	"video/x-motion-jpeg": [
		".mjpg"
	],
	"video/x-ms-asf": [
		".asf"
	],
	"video/x-ms-wm": [
		".wm"
	],
	"video/x-ms-wmv": [
		".wmv"
	],
	"video/x-ms-wmx": [
		".wmx"
	],
	"video/x-ms-wvx": [
		".wvx"
	],
	"video/x-msvideo": [
		".avi"
	],
	"video/x-qtc": [
		".qtc"
	],
	"video/x-sgi-movie": [
		".movie",
		".mv"
	],
	"x-conference/x-cooltalk": [
		".ice"
	],
	"x-epoc/x-sisx-app": [
		".sisx"
	],
	"x-world/x-3dmf": [
		".3dmf",
		".3dm",
		".qd3",
		".qd3d"
	],
	"x-world/x-vrml": [
		".vrm"
	],
	"x-world/x-vrt": [
		".vrt"
	],
	"xgl/drawing": [
		".xgz"
	],
	"xgl/movie": [
		".xmz"
	],
	"application/x-subrip":[
		".srt"
	]
}`
