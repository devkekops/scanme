var app = new Vue({
    el: '#app',
    data: {
      connection: null,
      isCheckAll: false,
      domainList: [],
      templateList: [],
      results: [],
      checkedTemplates: [],
      scanTemplates: [],
      btnText: "Выделить все",
      scanStatus: "",
      searchStatus: ""
    },
    mounted() {
        axios
          .get('http://0.0.0.0:8080/api/getTemplates')
          .then(response => (this.templateList = response.data));
      },
    methods: {
        search: function() {
            this.connection.send(JSON.stringify({event: "search", msg: this.domainList}));
        },
        checkAll: function(){
            this.isCheckAll = !this.isCheckAll;
            this.checkedTemplates = [];
            this.scanTemplates = [];
            if(this.isCheckAll){	// Check all
                this.btnText = "Убрать выделение"
                for (var i in this.templateList) {
                    this.scanTemplates.push(this.templateList[i]);
                }
            }
            else {
                this.btnText = "Выделить все"
            }
          },
          updateCheckall: function(){
            if(this.scanTemplates.length == this.scanTemplates.length){
                this.isCheckAll = true;
            }else{
                this.isCheckAll = false;
            }
        },
        scan: function() {
            // for (var i = 0; i < this.checkedTemplates.length; i++) {
            //     if (this.checkedTemplates[i].checked === true) {
            //         scanTemplates.push(this.checkedTemplates[i])
            //     }
            // }
            this.results = [];
            this.connection.send(JSON.stringify({event: "scan", msg: {domains: this.domainList, templates: this.scanTemplates}}));
        }
    },
    created: function() {
        this.$http.get(`http://0.0.0.0:8080/api/getDomains`, this.domainList).then(resp => {
            this.domainList = resp.data
        }).catch((resp) => {
            console.log('err:', resp)
        });

        this.connection = new WebSocket("ws://0.0.0.0:8080/ws");

        this.connection.onmessage = (event) => {
            //console.log(event.data);
            const obj = JSON.parse(event.data);
            //console.log(obj.event);
            //console.log(obj.msg);
            if (String(obj.event) === "search") {
                if (obj.msg === "start" || obj.msg == "finish") {
                    this.searchStatus = obj.msg
                }
                else {
                    this.domainList = this.domainList.concat(obj.msg)
                }
            } else if (String(obj.event) === "scan") {
                if (obj.msg === "start" || obj.msg == "finish") {
                    this.scanStatus = obj.msg
                }
                else {
                    this.results.push(obj.msg)
                    console.log(this.results)
                    console.log(this.results.length)
                }
            }
        };
    }
})
