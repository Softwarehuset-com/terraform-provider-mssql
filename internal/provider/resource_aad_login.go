package provider

//import (
//	"context"
//	"github.com/hashicorp/terraform-plugin-framework/resource"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//	"github.com/pkg/errors"
//	"github.com/softwarehuset/mssql/internal/model"
//)
//
//const AadLoginNameProp = "aad_login_name"
//
//type AadLoginConnector interface {
//	CreateAadLogin(ctx context.Context, name, defaultDatabase, defaultLanguage string) error
//	GetAadLogin(ctx context.Context, name string) (*model.AadLogin, error)
//	DeleteAadLogin(ctx context.Context, name string) error
//}
//
//func resourceAadLogin() resource.Resource {
//
//	return schema.Resource{
//		CreateContext: resourceAadLoginCreate,
//		ReadContext:   resourceAadLoginRead,
//		DeleteContext: resourceAadLoginDelete,
//		Importer: &schema.ResourceImporter{
//			StateContext: resourceAadLoginImport,
//		},
//		Schema: map[string]*schema.Schema{
//			serverProp: {
//				Type:     schema.TypeList,
//				MaxItems: 1,
//				Required: true,
//				ForceNew: true,
//				Elem: &schema.Resource{
//					Schema: getServerSchema(serverProp),
//				},
//			},
//			AadLoginNameProp: {
//				Type:     schema.TypeString,
//				Required: true,
//				ForceNew: true,
//			},
//			defaultDatabaseProp: {
//				Type:     schema.TypeString,
//				Optional: true,
//				ForceNew: true,
//				Default:  defaultDatabaseDefault,
//				DiffSuppressFunc: func(k, old, new string, data *schema.ResourceData) bool {
//					return (old == "" && new == defaultDatabaseDefault) || (old == defaultDatabaseDefault && new == "")
//				},
//			},
//			defaultLanguageProp: {
//				Type:     schema.TypeString,
//				Optional: true,
//				ForceNew: true,
//				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
//					return (old == "" && new == "us_english") || (old == "us_english" && new == "")
//				},
//			},
//		},
//		Timeouts: &schema.ResourceTimeout{
//			Default: defaulttime,
//		},
//	}
//}
//
//func resourceAadLoginCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
//
//	AadLoginName := data.Get(AadLoginNameProp).(string)
//	defaultDatabase := data.Get(defaultDatabaseProp).(string)
//	defaultLanguage := data.Get(defaultLanguageProp).(string)
//
//	connector, err := GetSqlConnector(meta, data)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	if err = connector.CreateAadLogin(ctx, AadLoginName, defaultDatabase, defaultLanguage); err != nil {
//		return diag.FromErr(errors.Wrapf(err, "unable to create AadLogin [%s]", AadLoginName))
//	}
//
//	return resourceAadLoginRead(ctx, data, meta)
//}
//
//func resourceAadLoginRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
//
//	AadLoginName := data.Get(AadLoginNameProp).(string)
//
//	connector, err := GetSqlConnector(meta, data)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	AadLogin, err := connector.GetAadLogin(ctx, AadLoginName)
//	if err != nil {
//		return diag.FromErr(errors.Wrapf(err, "unable to read AadLogin [%s]", AadLoginName))
//	}
//	if AadLogin == nil {
//		data.SetId("")
//	} else {
//		if err = data.Set(defaultDatabaseProp, AadLogin.DefaultDatabase); err != nil {
//			return diag.FromErr(err)
//		}
//		if err = data.Set(defaultLanguageProp, AadLogin.DefaultLanguage); err != nil {
//			return diag.FromErr(err)
//		}
//	}
//
//	return nil
//}
//
//func resourceAadLoginDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
//
//	AadLoginName := data.Get(AadLoginNameProp).(string)
//
//	connector, err := GetSqlConnector(meta, data)
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	if err = connector.DeleteAadLogin(ctx, AadLoginName); err != nil {
//		return diag.FromErr(errors.Wrapf(err, "unable to delete AadLogin [%s]", AadLoginName))
//	}
//
//	// d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
//	data.SetId("")
//
//	return nil
//}
//
//func resourceAadLoginImport(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
//
//	AadLoginName := data.Get(AadLoginNameProp).(string)
//
//	connector, err := GetSqlConnector(meta, data)
//	if err != nil {
//		return nil, err
//	}
//
//	AadLogin, err := connector.GetAadLogin(ctx, AadLoginName)
//	if err != nil {
//		return nil, errors.Wrapf(err, "unable to read AadLogin [%s] for import", AadLoginName)
//	}
//
//	if AadLogin == nil {
//		return nil, errors.Errorf("no AadLogin [%s] found for import", AadLoginName)
//	}
//
//	if err = data.Set(defaultDatabaseProp, AadLogin.DefaultDatabase); err != nil {
//		return nil, err
//	}
//	if err = data.Set(defaultLanguageProp, AadLogin.DefaultLanguage); err != nil {
//		return nil, err
//	}
//
//	return []*schema.ResourceData{data}, nil
//}
//
//func GetSqlConnector(meta interface{}, data *schema.ResourceData) (AadLoginConnector, error) {
//	provider := meta.(model.Provider)
//	connector, err := provider.GetSqlClient(serverProp, data)
//	if err != nil {
//		return nil, err
//	}
//	return connector.(AadLoginConnector), nil
//}
