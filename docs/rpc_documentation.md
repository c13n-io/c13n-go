# RPC API

## Contents

  
  ### [rpc/services/rpc.proto](#rpc/services/rpc.proto)
  <details>
  <summary>Messages</summary>

  
  <li>
    <a href="#services.AddContactRequest">AddContactRequest</a>
  </li>
  
  <li>
    <a href="#services.AddContactResponse">AddContactResponse</a>
  </li>
  
  <li>
    <a href="#services.AddDiscussionRequest">AddDiscussionRequest</a>
  </li>
  
  <li>
    <a href="#services.AddDiscussionResponse">AddDiscussionResponse</a>
  </li>
  
  <li>
    <a href="#services.Chain">Chain</a>
  </li>
  
  <li>
    <a href="#services.ConnectNodeRequest">ConnectNodeRequest</a>
  </li>
  
  <li>
    <a href="#services.ConnectNodeResponse">ConnectNodeResponse</a>
  </li>
  
  <li>
    <a href="#services.ContactInfo">ContactInfo</a>
  </li>
  
  <li>
    <a href="#services.DiscussionInfo">DiscussionInfo</a>
  </li>
  
  <li>
    <a href="#services.DiscussionOptions">DiscussionOptions</a>
  </li>
  
  <li>
    <a href="#services.EstimateMessageRequest">EstimateMessageRequest</a>
  </li>
  
  <li>
    <a href="#services.EstimateMessageResponse">EstimateMessageResponse</a>
  </li>
  
  <li>
    <a href="#services.GetContactsRequest">GetContactsRequest</a>
  </li>
  
  <li>
    <a href="#services.GetContactsResponse">GetContactsResponse</a>
  </li>
  
  <li>
    <a href="#services.GetDiscussionHistoryByIDRequest">GetDiscussionHistoryByIDRequest</a>
  </li>
  
  <li>
    <a href="#services.GetDiscussionHistoryResponse">GetDiscussionHistoryResponse</a>
  </li>
  
  <li>
    <a href="#services.GetDiscussionStatisticsRequest">GetDiscussionStatisticsRequest</a>
  </li>
  
  <li>
    <a href="#services.GetDiscussionStatisticsResponse">GetDiscussionStatisticsResponse</a>
  </li>
  
  <li>
    <a href="#services.GetDiscussionsRequest">GetDiscussionsRequest</a>
  </li>
  
  <li>
    <a href="#services.GetDiscussionsResponse">GetDiscussionsResponse</a>
  </li>
  
  <li>
    <a href="#services.GetNodesRequest">GetNodesRequest</a>
  </li>
  
  <li>
    <a href="#services.Message">Message</a>
  </li>
  
  <li>
    <a href="#services.MessageOptions">MessageOptions</a>
  </li>
  
  <li>
    <a href="#services.NodeInfo">NodeInfo</a>
  </li>
  
  <li>
    <a href="#services.NodeInfoResponse">NodeInfoResponse</a>
  </li>
  
  <li>
    <a href="#services.OpenChannelRequest">OpenChannelRequest</a>
  </li>
  
  <li>
    <a href="#services.OpenChannelResponse">OpenChannelResponse</a>
  </li>
  
  <li>
    <a href="#services.PageOptions">PageOptions</a>
  </li>
  
  <li>
    <a href="#services.PaymentHop">PaymentHop</a>
  </li>
  
  <li>
    <a href="#services.PaymentRoute">PaymentRoute</a>
  </li>
  
  <li>
    <a href="#services.RemoveContactByAddressRequest">RemoveContactByAddressRequest</a>
  </li>
  
  <li>
    <a href="#services.RemoveContactByIDRequest">RemoveContactByIDRequest</a>
  </li>
  
  <li>
    <a href="#services.RemoveContactResponse">RemoveContactResponse</a>
  </li>
  
  <li>
    <a href="#services.RemoveDiscussionRequest">RemoveDiscussionRequest</a>
  </li>
  
  <li>
    <a href="#services.RemoveDiscussionResponse">RemoveDiscussionResponse</a>
  </li>
  
  <li>
    <a href="#services.SearchNodeByAddressRequest">SearchNodeByAddressRequest</a>
  </li>
  
  <li>
    <a href="#services.SearchNodeByAliasRequest">SearchNodeByAliasRequest</a>
  </li>
  
  <li>
    <a href="#services.SelfBalanceRequest">SelfBalanceRequest</a>
  </li>
  
  <li>
    <a href="#services.SelfBalanceResponse">SelfBalanceResponse</a>
  </li>
  
  <li>
    <a href="#services.SelfInfoRequest">SelfInfoRequest</a>
  </li>
  
  <li>
    <a href="#services.SelfInfoResponse">SelfInfoResponse</a>
  </li>
  
  <li>
    <a href="#services.SendMessageRequest">SendMessageRequest</a>
  </li>
  
  <li>
    <a href="#services.SendMessageResponse">SendMessageResponse</a>
  </li>
  
  <li>
    <a href="#services.SubscribeMessageRequest">SubscribeMessageRequest</a>
  </li>
  
  <li>
    <a href="#services.SubscribeMessageResponse">SubscribeMessageResponse</a>
  </li>
  
  <li>
    <a href="#services.UpdateDiscussionLastReadRequest">UpdateDiscussionLastReadRequest</a>
  </li>
  
  <li>
    <a href="#services.UpdateDiscussionResponse">UpdateDiscussionResponse</a>
  </li>
  
  <li>
    <a href="#services.Version">Version</a>
  </li>
  
  <li>
    <a href="#services.VersionRequest">VersionRequest</a>
  </li>
  
  
  
  </details>

  <details>
  <summary>Services</summary>

  
  <li>
    <a href="#services.ChannelService">ChannelService</a>
  </li>
  
  <li>
    <a href="#services.ContactService">ContactService</a>
  </li>
  
  <li>
    <a href="#services.DiscussionService">DiscussionService</a>
  </li>
  
  <li>
    <a href="#services.MessageService">MessageService</a>
  </li>
  
  <li>
    <a href="#services.NodeInfoService">NodeInfoService</a>
  </li>
  
  </details>


### [Scalar Value Types](#scalar-value-types)




<h2 id="rpc/services/rpc.proto">rpc/services/rpc.proto</h2>



<h3 id="services.AddContactRequest">AddContactRequest</h3>
Corresponds to a request to add a node as a contact.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>contact</td>
      <td><a href="#services.ContactInfo">ContactInfo</a></td>
      <td></td>
      <td>
        <p>The node to add as a contact.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>contact</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.AddContactResponse">AddContactResponse</h3>
A AddContactResponse is received in response to an AddContact rpc call.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>contact</td>
      <td><a href="#services.ContactInfo">ContactInfo</a></td>
      <td></td>
      <td>
        <p>The newly added contact's information.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>contact</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.AddDiscussionRequest">AddDiscussionRequest</h3>
Corresponds to a request to add a discussion to database.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>discussion</td>
      <td><a href="#services.DiscussionInfo">DiscussionInfo</a></td>
      <td></td>
      <td>
        <p>  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>discussion</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.AddDiscussionResponse">AddDiscussionResponse</h3>
An AddDiscussionResponse is received in response to an AddDiscussion rpc call.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>discussion</td>
      <td><a href="#services.DiscussionInfo">DiscussionInfo</a></td>
      <td></td>
      <td>
        <p>  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.Chain">Chain</h3>
Represents a blockchain and network for a Lightning node.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>chain</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The blockchain in use.  </p>
      </td>
    </tr>
    
    <tr>
      <td>network</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The network a node is operating on.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.ConnectNodeRequest">ConnectNodeRequest</h3>
Corresponds to a request to create a peer connection with a node.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>address</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The address of the node to connect.  </p>
      </td>
    </tr>
    
    <tr>
      <td>hostport</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The network location of the node.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>address</td>
        <td>
          <ul>
            
              
                <li>regex: ^[a-z0-9]{66}$</li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.ConnectNodeResponse">ConnectNodeResponse</h3>
A ConnectNodeResponse is received in response to a ConnectNode request.




<h3 id="services.ContactInfo">ContactInfo</h3>
A message representing a contact of the application.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>node</td>
      <td><a href="#services.NodeInfo">NodeInfo</a></td>
      <td></td>
      <td>
        <p>The node corresponding to the contact.  </p>
      </td>
    </tr>
    
    <tr>
      <td>id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The contact id.  </p>
      </td>
    </tr>
    
    <tr>
      <td>display_name</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>A contact's chat nickname.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>node</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
      <tr>
        <td>id</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.DiscussionInfo">DiscussionInfo</h3>
Represents the information for a specific discussion.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The discussion id.  </p>
      </td>
    </tr>
    
    <tr>
      <td>participants</td>
      <td><a href="#string">string</a></td>
      <td>repeated</td>
      <td>
        <p>The list of participants in the discussion.  </p>
      </td>
    </tr>
    
    <tr>
      <td>options</td>
      <td><a href="#services.DiscussionOptions">DiscussionOptions</a></td>
      <td></td>
      <td>
        <p>  </p>
      </td>
    </tr>
    
    <tr>
      <td>last_read_msg_id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The id of the last read message in the discussion.  </p>
      </td>
    </tr>
    
    <tr>
      <td>last_msg_id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The id of the last discussion message.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>id</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
      <tr>
        <td>participants</td>
        <td>
          <ul>
            
              
                <li>repeated_count_min: 1</li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.DiscussionOptions">DiscussionOptions</h3>
DiscussionOptions represents the per-discussion options.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>fee_limit_msat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The maximum fee allowed for sending a message (in millisatoshi).<br>If not set, the default fee limit, as defined in the app package, is used.  </p>
      </td>
    </tr>
    
    <tr>
      <td>anonymous</td>
      <td><a href="#bool">bool</a></td>
      <td></td>
      <td>
        <p>Whether to send as anonymous on this discussion.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.EstimateMessageRequest">EstimateMessageRequest</h3>
Corresponds to a request to estimate a message.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>discussion_id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The discussion id where the message is to be sent.  </p>
      </td>
    </tr>
    
    <tr>
      <td>payload</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The message payload (as a string).  </p>
      </td>
    </tr>
    
    <tr>
      <td>amt_msat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The intended payment amount to the recipient of the message (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>options</td>
      <td><a href="#services.MessageOptions">MessageOptions</a></td>
      <td></td>
      <td>
        <p>The message option overrides for the current message.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>amt_msat</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.EstimateMessageResponse">EstimateMessageResponse</h3>
A EstimateMessageResponse is received in response to a EstimateMessage rpc call.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>success_prob</td>
      <td><a href="#double">double</a></td>
      <td></td>
      <td>
        <p>The probability of successful arrival of the message,
as reported by the Lightning daemon's mission control.  </p>
      </td>
    </tr>
    
    <tr>
      <td>message</td>
      <td><a href="#services.Message">Message</a></td>
      <td></td>
      <td>
        <p>Contains the estimated route and fees for the requested message.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>success_prob</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
      <tr>
        <td>message</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.GetContactsRequest">GetContactsRequest</h3>
Corresponds to a request to list all contacts.




<h3 id="services.GetContactsResponse">GetContactsResponse</h3>
A GetContactsResponse is received in response to a GetContacts rpc call.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>contacts</td>
      <td><a href="#services.ContactInfo">ContactInfo</a></td>
      <td>repeated</td>
      <td>
        <p>The list of contacts in the database.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.GetDiscussionHistoryByIDRequest">GetDiscussionHistoryByIDRequest</h3>
Corresponds to a request to create a stream over which to receive
previously exchanged messages of the identified discussion.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The discussion id of interest.  </p>
      </td>
    </tr>
    
    <tr>
      <td>page_options</td>
      <td><a href="#services.PageOptions">PageOptions</a></td>
      <td></td>
      <td>
        <p>  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>id</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.GetDiscussionHistoryResponse">GetDiscussionHistoryResponse</h3>
A GetDiscussionHistoryResponse is received in response to
a GetHistory rpc call, and represents an exchanged message.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>message</td>
      <td><a href="#services.Message">Message</a></td>
      <td></td>
      <td>
        <p>The exchanged message.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.GetDiscussionStatisticsRequest">GetDiscussionStatisticsRequest</h3>
Corresponds to a request for statistics about the requested
discussion, identified by its id.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The discussion id.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.GetDiscussionStatisticsResponse">GetDiscussionStatisticsResponse</h3>
A GetDiscussionStatisticsResponse is received in response to
a GetDiscussionStatistics rpc call.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>amt_msat_sent</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The total amount sent in the discussion (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>amt_msat_received</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The total amount received in the discussion (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>amt_msat_fees</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The total amount of fees for sent messages in the discussion (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>messages_sent</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The total amount of sent messages in the discussion.  </p>
      </td>
    </tr>
    
    <tr>
      <td>messages_received</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The total amount of received messages in the discussion.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.GetDiscussionsRequest">GetDiscussionsRequest</h3>
Corresponds to a request to receive all discussion info.




<h3 id="services.GetDiscussionsResponse">GetDiscussionsResponse</h3>
A GetDiscussionsResponse is received in the stream returned in response
to a GetDiscussions rpc call, and represents a discussion.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>discussion</td>
      <td><a href="#services.DiscussionInfo">DiscussionInfo</a></td>
      <td></td>
      <td>
        <p>  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>discussion</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.GetNodesRequest">GetNodesRequest</h3>
Corresponds to a request to list all nodes on the Lightning Network.




<h3 id="services.Message">Message</h3>
A message representing a message of the application.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The unique id of the message.  </p>
      </td>
    </tr>
    
    <tr>
      <td>discussion_id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The discussion id this message is associated with.  </p>
      </td>
    </tr>
    
    <tr>
      <td>sender</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The Lightning address of the sender node.  </p>
      </td>
    </tr>
    
    <tr>
      <td>receiver</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The Lightning address of the receiver node.  </p>
      </td>
    </tr>
    
    <tr>
      <td>sender_verified</td>
      <td><a href="#bool">bool</a></td>
      <td></td>
      <td>
        <p>Whether the message sender was verified.  </p>
      </td>
    </tr>
    
    <tr>
      <td>payload</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The message payload.  </p>
      </td>
    </tr>
    
    <tr>
      <td>amt_msat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The amount paid over this message (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>total_fees_msat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The total routing fees paid for this message across all routes (in millisatoshi).<br>This field is meaningful only for sent and estimated messages.  </p>
      </td>
    </tr>
    
    <tr>
      <td>sent_timestamp</td>
      <td><a href="#google.protobuf.Timestamp">google.protobuf.Timestamp</a></td>
      <td></td>
      <td>
        <p>The time the message was sent.  </p>
      </td>
    </tr>
    
    <tr>
      <td>received_timestamp</td>
      <td><a href="#google.protobuf.Timestamp">google.protobuf.Timestamp</a></td>
      <td></td>
      <td>
        <p>The time the message was received.  </p>
      </td>
    </tr>
    
    <tr>
      <td>payment_routes</td>
      <td><a href="#services.PaymentRoute">PaymentRoute</a></td>
      <td>repeated</td>
      <td>
        <p>The routes that fulfilled this message.<br>This field is meaningful only for sent and estimated messages.  </p>
      </td>
    </tr>
    
    <tr>
      <td>preimage</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The preimage belonging to the associated payment.<br>This field is only meaningful for received messages and
messages sent successfully to non-group discussions.  </p>
      </td>
    </tr>
    
    <tr>
      <td>pay_req</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The payment request this message was paid to.<br>If empty, corresponds to a spontaneous payment.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>sender</td>
        <td>
          <ul>
            
              
                <li>regex: ^[a-z0-9]{66}$</li>
              
            
          </ul>
        </td>
      </tr>
      
      <tr>
        <td>receiver</td>
        <td>
          <ul>
            
              
                <li>regex: ^[a-z0-9]{66}$</li>
              
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
      <tr>
        <td>amt_msat</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
      <tr>
        <td>total_fees_msat</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.MessageOptions">MessageOptions</h3>
MessageOptions represents messaging options.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>fee_limit_msat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The maximum fee allowed for a message (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>anonymous</td>
      <td><a href="#bool">bool</a></td>
      <td></td>
      <td>
        <p>Whether to include the sender address when sending a message.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.NodeInfo">NodeInfo</h3>
A message representing a node on the Lightning network.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>alias</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>A node's Lightning alias.  </p>
      </td>
    </tr>
    
    <tr>
      <td>address</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>A node's Lightning address.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>address</td>
        <td>
          <ul>
            
              
                <li>regex: ^[a-z0-9]{66}$</li>
              
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.NodeInfoResponse">NodeInfoResponse</h3>
A NodeInfoResponse is received in response to a node query.<br>It contains all visible nodes corresponding to the query.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>nodes</td>
      <td><a href="#services.NodeInfo">NodeInfo</a></td>
      <td>repeated</td>
      <td>
        <p>The list of Lightning nodes matching the query.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.OpenChannelRequest">OpenChannelRequest</h3>
Corresponds to a request to open a channel.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>address</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The address of the node to open a channel to.  </p>
      </td>
    </tr>
    
    <tr>
      <td>amt_msat</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The total amount to be committed to the channel (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>push_amt_msat</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The amount to be sent to the other party (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>min_input_confs</td>
      <td><a href="#uint32">uint32</a></td>
      <td></td>
      <td>
        <p>The minimum number of confirmations each input
of the channel funding transaction must have.<br>In case of 0, unconfirmed funds can be used.  </p>
      </td>
    </tr>
    
    <tr>
      <td>target_confirmation_block</td>
      <td><a href="#uint32">uint32</a></td>
      <td></td>
      <td>
        <p>The number of blocks the funding transaction should confirm by.
This is used for fee estimation.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.OpenChannelResponse">OpenChannelResponse</h3>
An OpenChannelResponse is received in response to an OpenChannel call.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>funding_txid</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The channel funding transaction.  </p>
      </td>
    </tr>
    
    <tr>
      <td>output_index</td>
      <td><a href="#uint32">uint32</a></td>
      <td></td>
      <td>
        <p>The output index of the funding transaction.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.PageOptions">PageOptions</h3>
Corresponds to pagination parameters for requests.
Represents a request for page_size elements,
skipping the most recent skip_recent elements (reverse order).


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>skip_recent</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The number of most recent elements to skip.  </p>
      </td>
    </tr>
    
    <tr>
      <td>page_size</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The number of elements to return.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>skip_recent</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
      <tr>
        <td>page_size</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.PaymentHop">PaymentHop</h3>
A message representing a hop of the route for sending a message.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>chan_id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The channel id.  </p>
      </td>
    </tr>
    
    <tr>
      <td>hop_address</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The address of the hop node.  </p>
      </td>
    </tr>
    
    <tr>
      <td>amt_to_forward_msat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The amount to be forwarded by the hop (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>fee_msat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The fee to be paid to the hop for forwarding the message (in millisatoshi).  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>hop_address</td>
        <td>
          <ul>
            
              
                <li>regex: ^[a-z0-9]{66}$</li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.PaymentRoute">PaymentRoute</h3>
A message representing a route fulfilling a payment HTLC.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>hops</td>
      <td><a href="#services.PaymentHop">PaymentHop</a></td>
      <td>repeated</td>
      <td>
        <p>The list of hops for this route.  </p>
      </td>
    </tr>
    
    <tr>
      <td>total_timelock</td>
      <td><a href="#uint32">uint32</a></td>
      <td></td>
      <td>
        <p>The total timelock of the route.  </p>
      </td>
    </tr>
    
    <tr>
      <td>route_amt_msat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The amount sent via this route, disregarding the route fees (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>route_fees_msat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The fees paid for this route (in millisatoshi).  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.RemoveContactByAddressRequest">RemoveContactByAddressRequest</h3>
Corresponds to a request to remove a contact.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>address</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The Lightning address of the contact to remove.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>address</td>
        <td>
          <ul>
            
              
                <li>regex: ^[a-z0-9]{66}$</li>
              
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.RemoveContactByIDRequest">RemoveContactByIDRequest</h3>
Corresponds to a request to remove a contact.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The id of the contact to remove.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>id</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.RemoveContactResponse">RemoveContactResponse</h3>
A RemoveContactResponse is received in response to a RemoveContactBy* rpc call.




<h3 id="services.RemoveDiscussionRequest">RemoveDiscussionRequest</h3>
Corresponds to a request to remove a discussion.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The id of the discussion to remove.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>id</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.RemoveDiscussionResponse">RemoveDiscussionResponse</h3>
A RemoveDiscussionResponse is received in response to a RemoveDiscussion rpc call.




<h3 id="services.SearchNodeByAddressRequest">SearchNodeByAddressRequest</h3>
Corresponds to a node query based on the node address.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>address</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The node address to search.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>address</td>
        <td>
          <ul>
            
              
                <li>regex: ^[a-z0-9]{66}$</li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.SearchNodeByAliasRequest">SearchNodeByAliasRequest</h3>
Corresponds to a node query based on the node alias.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>alias</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The node alias substring to search.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.SelfBalanceRequest">SelfBalanceRequest</h3>
Corresponds to a query to retrieve the balance of the underlying lightning node.




<h3 id="services.SelfBalanceResponse">SelfBalanceResponse</h3>
A SelfBalanceResponse is received in response to a GetSelfBalance rpc call.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>wallet_confirmed_sat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The confirmed balance of the node's wallet (in satoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>wallet_unconfirmed_sat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The unconfirmed balance of the node's wallet (in satoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>channel_local_msat</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The local balance available across all open channels (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>channel_remote_msat</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The remote balance available across all open channels (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>pending_open_local_msat</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The local balance in pending open channels (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>pending_open_remote_msat</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The remote balance in pending open channels (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>unsettled_local_msat</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The local balance unsettled across all open channels (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>unsettled_remote_msat</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The remote balance unsettled across all open channels (in millisatoshi).  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.SelfInfoRequest">SelfInfoRequest</h3>
Correponds to a query to retrieve the information of the underlying lightning node.




<h3 id="services.SelfInfoResponse">SelfInfoResponse</h3>
A SelfInfoResponse is received in response to a GetSelfInfo rpc call.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>info</td>
      <td><a href="#services.NodeInfo">NodeInfo</a></td>
      <td></td>
      <td>
        <p>General node information about the current node.  </p>
      </td>
    </tr>
    
    <tr>
      <td>chains</td>
      <td><a href="#services.Chain">Chain</a></td>
      <td>repeated</td>
      <td>
        <p>A list of chain networks the node is operating on.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>info</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.SendMessageRequest">SendMessageRequest</h3>
Corresponds to a request to send a message.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>discussion_id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The discussion id where the message is to be sent.  </p>
      </td>
    </tr>
    
    <tr>
      <td>payload</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The message payload (as a string).  </p>
      </td>
    </tr>
    
    <tr>
      <td>amt_msat</td>
      <td><a href="#int64">int64</a></td>
      <td></td>
      <td>
        <p>The intended payment amount to the recipient of the message (in millisatoshi).  </p>
      </td>
    </tr>
    
    <tr>
      <td>pay_req</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>A payment request (invoice) to pay to.<br>If empty, a spontaneous message is sent.
If specified, discussion_id is not used and should not be specified.
Instead, the message will be sent to the discussion associated with
the recipient specified by the the payment request (which will be
created if it does not exist).
The discussion_id will be returned in the response.  </p>
      </td>
    </tr>
    
    <tr>
      <td>options</td>
      <td><a href="#services.MessageOptions">MessageOptions</a></td>
      <td></td>
      <td>
        <p>The message option overrides for the current message.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>amt_msat</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.SendMessageResponse">SendMessageResponse</h3>
A SendMessageResponse is received in response to a SendMessage rpc call.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>sent_message</td>
      <td><a href="#services.Message">Message</a></td>
      <td></td>
      <td>
        <p>The sent message.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>sent_message</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.SubscribeMessageRequest">SubscribeMessageRequest</h3>
Corresponds to a request to create a stream
over which to be notified of received messages.




<h3 id="services.SubscribeMessageResponse">SubscribeMessageResponse</h3>
A SubscribeMessageResponse is received in the stream returned in response to
a SubscribeMessages rpc call, and represents a received message.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>received_message</td>
      <td><a href="#services.Message">Message</a></td>
      <td></td>
      <td>
        <p>The received message.  </p>
      </td>
    </tr>
    
  </tbody>
</table>


  
  
  <h4>Validated Fields</h4>
  <table>
    <thead>
      <tr>
        <td>Field</td>
        <td>Validations</td>
      </tr>
    </thead>
    <tbody>
      
      <tr>
        <td>received_message</td>
        <td>
          <ul>
            
              
                <li><b>Required</b></li>
              
            
          </ul>
        </td>
      </tr>
      
    </tbody>
  </table>
  
  


<h3 id="services.UpdateDiscussionLastReadRequest">UpdateDiscussionLastReadRequest</h3>
Represents a request to update the last read discussion message.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>discussion_id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The discussion id.  </p>
      </td>
    </tr>
    
    <tr>
      <td>last_read_msg_id</td>
      <td><a href="#uint64">uint64</a></td>
      <td></td>
      <td>
        <p>The message id to mark as the last read.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.UpdateDiscussionResponse">UpdateDiscussionResponse</h3>
An UpdateDiscussionResponse is returned in reponse
to an UpdateDiscussionLastRead request.




<h3 id="services.Version">Version</h3>
A message containing the current c13n version and build information.


<table class="field-table">
  <thead>
    <tr>
      <td>Field</td>
      <td>Type</td>
      <td>Label</td>
      <td>Description</td>
    </tr>
  </thead>
  <tbody>
    
    <tr>
      <td>version</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The semantic version of c13n.  </p>
      </td>
    </tr>
    
    <tr>
      <td>commit</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The commit descriptor of the build.  </p>
      </td>
    </tr>
    
    <tr>
      <td>commit_hash</td>
      <td><a href="#string">string</a></td>
      <td></td>
      <td>
        <p>The commit hash of the build.  </p>
      </td>
    </tr>
    
  </tbody>
</table>




<h3 id="services.VersionRequest">VersionRequest</h3>





